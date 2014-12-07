package prov

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"
)

func Include(dir string, vars, args map[interface{}]interface{}, live bool) (ok, changed int, err error) {
	start := time.Now()
	path, pathOK := getStringVar(args, "path")
	if !pathOK {
		err = ErrInvalidArg("path")
		return
	}
	vars = copyVars(vars)
	newVars, newVarsOK := getVarsVar(args, "vars")
	if newVarsOK {
		setVars(vars, newVars)
	}
	filename := filepath.Join(dir, path)
	newDir := filepath.Dir(filename)
	log.Printf(">> %q.\n", filename)
	tasks, err := readFile(filename, vars)
	if err != nil {
		return
	}
	changedTasks := map[interface{}]struct{}{}
	for _, task := range tasks {
		name, nameOK := getStringVar(task, "name")
		taskType, _ := getStringVar(task, "task")
		if !nameOK {
			name = taskType
		}
		when, whenOK := task["when"]
		if whenOK {
			var check bool
			check, err = checkWhen(when, changedTasks)
			if err != nil {
				return
			}
			if !check {
				log.Printf("-- %v\nSKIPPED\n", name)
				continue
			}
		}
		if taskType == "include" {
			var innerOK, innerChanged int
			innerOK, innerChanged, err = Include(newDir, vars, task, live)
			if err != nil {
				return
			}
			ok += innerOK
			changed += innerChanged
			if innerChanged > 0 && nameOK {
				changedTasks[name] = struct{}{}
			}
		} else {
			var status Status
			status, err = RunTask(newDir, name, vars, task, live)
			if err != nil {
				return
			}
			switch status {
			case OK:
				ok++
			case Changed:
				changed++
				if nameOK {
					changedTasks[name] = struct{}{}
				}
			}
		}
	}
	log.Printf("<< %q (%s).\n", filename, time.Since(start).String())
	return
}

var startMarkerRE = regexp.MustCompile(`(?m)^---$`)
var endMarkerRE = regexp.MustCompile(`(?m)^\.\.\.$`)

func readFile(filename string, vars map[interface{}]interface{}) (tasks []map[interface{}]interface{}, err error) {
	fileRaw, err := ioutil.ReadFile(filename)
	if err != nil {
		return
	}
	if len(startMarkerRE.FindAllIndex(fileRaw, -1)) > 1 {
		var newVars map[interface{}]interface{}
		err = yaml.Unmarshal([]byte(fileRaw), &newVars)
		if err != nil {
			return
		}
		setVars(vars, newVars)
		fileRaw = []byte(endMarkerRE.Split(string(fileRaw), 2)[1])
	}
	templ, err := template.New("template").Parse(string(fileRaw))
	if err != nil {
		return
	}
	buffer := bytes.NewBuffer(nil)
	err = templ.Execute(buffer, vars)
	if err != nil {
		return
	}
	err = yaml.Unmarshal([]byte(strings.Replace(buffer.String(), "<no value>", "", -1)), &tasks)
	if err != nil {
		return
	}
	return
}

func checkWhen(when interface{}, changedTasks map[interface{}]struct{}) (bool, error) {
	list, ok := when.([]interface{})
	if !ok {
		return false, errors.New("when is not an expected list")
	}
	result := false
	for _, e := range list {
		m, ok := e.(map[interface{}]interface{})
		if !ok {
			return false, errors.New("when item is not an expected map")
		}
		if len(m) != 1 {
			return false, errors.New("when item is not an expected map with exactly one element")
		}
		var k, v interface{}
		for k, v = range m {
		}
		nextResult, err := checkCond(k, v, changedTasks)
		if err != nil {
			return false, err
		}
		result = result || nextResult
	}
	return result, nil
}

func checkCond(check, value interface{}, changedTasks map[interface{}]struct{}) (bool, error) {
	switch check {
	case "ok":
		_, ok := changedTasks[value]
		return !ok, nil
	case "changed":
		_, ok := changedTasks[value]
		return ok, nil
	default:
		return false, fmt.Errorf("unrecognized when check type: %v", check)
	}
}
