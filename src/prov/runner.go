package prov

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
	"text/template"
	"time"

	"gopkg.in/yaml.v2"
)

type ErrInvalidArg string

func (me ErrInvalidArg) Error() string {
	return fmt.Sprintf("missing or invalid argument: %q", string(me))
}

type ErrCommandFailed struct {
	Err    error
	Output []byte
}

func (me ErrCommandFailed) Error() string {
	return fmt.Sprintf("%s: %s", me.Err.Error(), me.Output)
}

type ErrTaskFailed struct {
	Key interface{}
	Err error
}

func (me ErrTaskFailed) Error() string {
	return fmt.Sprintf("task %v failed: %s", me.Key, me.Err.Error())
}

type TaskRunner func(dir string, vars, args map[interface{}]interface{}, live bool) (changed bool, err error)

var taskRunners = map[string]TaskRunner{}

func RegisterRunner(name string, runner TaskRunner) {
	taskRunners[name] = runner
}

func RunFile(filename string, vars map[interface{}]interface{}, live bool) (ok, changed, skipped int, err error) {
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return
	}
	log.Printf("running %q", absFilename)
	tasks, err := readTasks(absFilename, vars)
	if err != nil {
		return
	}
	dir := filepath.Dir(absFilename)
	changedTasks := map[interface{}]struct{}{}
	for _, task := range tasks {
		start := time.Now()
		name, nameOK := task["name"]
		taskType, taskTypeOK := getStringVar(task, "task")
		if !taskTypeOK {
			err = ErrTaskFailed{name, ErrInvalidArg("task")}
			return
		}
		if !nameOK {
			name = taskType
		}
		log.Printf("======== %v", name)
		when, whenOK := task["when"]
		if whenOK {
			var check bool
			check, err = checkWhen(when, changedTasks)
			if err != nil {
				return
			}
			if !check {
				skipped++
				log.Printf("skipped")
				continue
			}
		}
		switch taskType {
		case "include":
			path, pathOK := getStringVar(task, "path")
			if !pathOK {
				err = ErrTaskFailed{name, ErrInvalidArg("path")}
				return
			}
			log.Printf(">>>>>>>> %s", path)
			vars = copyVars(vars)
			newVars, _ := getVarsVar(task, "vars")
			setVars(vars, newVars)
			filename := filepath.Join(dir, path)
			var innerOK, innerChanged, innerSkipped int
			innerOK, innerChanged, innerSkipped, err = RunFile(filename, vars, live)
			if err != nil {
				return
			}
			ok += innerOK
			changed += innerChanged
			skipped += innerSkipped
			if innerChanged > 0 && nameOK {
				changedTasks[name] = struct{}{}
			}
			log.Printf("<<<<<<<< %s", path)
			if innerChanged > 0 {
				log.Printf("CHANGED (%s)", time.Since(start).String())
			} else {
				log.Printf("ok (%s)", time.Since(start).String())
			}
		default:
			runner, runnerOK := taskRunners[taskType]
			if !runnerOK {
				err = ErrTaskFailed{name, fmt.Errorf("unrecognized task: %q", task)}
				return
			}
			taskChanged := false
			taskChanged, err = runner(dir, vars, task, live)
			if err != nil {
				err = ErrTaskFailed{name, err}
				return
			}
			if taskChanged {
				changed++
				if nameOK {
					changedTasks[name] = struct{}{}
				}
				log.Printf("CHANGED (%s)", time.Since(start).String())
			} else {
				ok++
				log.Printf("ok (%s)", time.Since(start).String())
			}
		}
	}
	return
}

var startMarkerRE = regexp.MustCompile(`(?m)^---$`)
var endMarkerRE = regexp.MustCompile(`(?m)^\.\.\.$`)

func readTasks(filename string, vars map[interface{}]interface{}) (tasks []map[interface{}]interface{}, err error) {
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
	templ, err := template.New("default").Parse(string(fileRaw))
	if err != nil {
		return
	}
	buffer := bytes.NewBuffer(nil)
	err = templ.Execute(buffer, vars)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(buffer.Bytes(), &tasks)
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
	var result bool
	switch check {
	case "ok":
		_, result = changedTasks[value]
		result = !result
	case "changed":
		_, result = changedTasks[value]
	default:
		return false, fmt.Errorf("unrecognized when check type: %v", check)
	}
	log.Printf("check %q on %q: %v", check, value, result)
	return result, nil
}
