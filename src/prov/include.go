package prov

import (
	"bytes"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"text/template"
	"time"

	"gopkg.in/yaml.v1"
)

func Include(dir string, vars Vars, args Args, run bool) (ok, changed int, err error) {
	start := time.Now()
	path, pathOK := args.String("path")
	if !pathOK {
		err = errors.New(`missing or invalid "path" argument`)
		return
	}
	filename := filepath.Join(dir, path)
	log.Printf("==> %q.\n\n", filename)
	tasks, handlers, err := readFile(filename, vars)
	queue := &TaskQueue{
		Tasks: tasks,
	}
	handlersByName := map[string]Args{}
	for _, handler := range handlers {
		name, nameOK := handler.Name()
		if !nameOK {
			err = errors.New(`handler missing or invalid "name" parameter`)
			return
		}
		_, exists := handlersByName[name]
		if exists {
			err = fmt.Errorf("duplicate handler with name %q", name)
			return
		}
		handlersByName[name] = handler
	}
	activatedHandlers := map[string]struct{}{}
	activateHandler := func(name string) error {
		handler, handlerOK := handlersByName[name]
		if !handlerOK {
			return fmt.Errorf("unrecognized handler %q", name)
		}
		_, handlerOK = activatedHandlers[name]
		if !handlerOK {
			activatedHandlers[name] = struct{}{}
			queue.Handlers = append(queue.Handlers, handler)
		}
		return nil
	}
	fileDir := filepath.Dir(filename)
	for task := queue.Next(); task != nil; task = queue.Next() {
		if isInclude(task) {
			var innerOK, innerChanged int
			innerOK, innerChanged, err = Include(fileDir, vars, task, run)
			if err != nil {
				return
			}
			ok += innerOK
			changed += innerChanged
			if innerChanged > 0 {
				handler, handlerOK := task.String("notify")
				if handlerOK {
					err = activateHandler(handler)
					if err != nil {
						return
					}
				}
			}
		} else if isHandlers(task) {
			queue.FlushHandlers = true
			activatedHandlers = map[string]struct{}{}
		} else {
			var status Status
			status, err = RunTask(fileDir, vars, task, run)
			if err != nil {
				return
			}
			switch status {
			case OK:
				ok++
			case Changed:
				changed++
				handler, handlerOK := task.String("notify")
				if handlerOK {
					err = activateHandler(handler)
					if err != nil {
						return
					}
				}
			}
		}
	}
	log.Printf("<== %q (%s).\n\n", filename, time.Since(start).String())
	return
}

func readFile(filename string, vars Vars) (tasks []Args, handlers []Args, err error) {
	fileRaw, err := ioutil.ReadFile(filename)
	if err != nil {
		return
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
	var contents struct {
		Vars     Vars
		Tasks    []Args
		Handlers []Args
	}
	err = yaml.Unmarshal(buffer.Bytes(), &contents)
	if err != nil {
		return
	}
	vars.SetVars(contents.Vars)
	contents.Vars = nil
	contents.Tasks = nil
	contents.Handlers = nil
	buffer.Reset()
	err = templ.Execute(buffer, vars)
	if err != nil {
		return
	}
	err = yaml.Unmarshal(buffer.Bytes(), &contents)
	if err != nil {
		return
	}
	tasks = contents.Tasks
	handlers = contents.Handlers
	return
}

type TaskQueue struct {
	Tasks         []Args
	tasksIndex    int
	Handlers      []Args
	handlersIndex int
	FlushHandlers bool
}

func (me *TaskQueue) Next() Args {
	if me.FlushHandlers {
		if me.handlersIndex < len(me.Handlers) {
			task := me.Handlers[me.handlersIndex]
			me.handlersIndex++
			return task
		} else {
			me.FlushHandlers = false
		}
	}
	if me.tasksIndex < len(me.Tasks) {
		task := me.Tasks[me.tasksIndex]
		me.tasksIndex++
		return task
	}
	if me.handlersIndex < len(me.Handlers) {
		task := me.Handlers[me.handlersIndex]
		me.handlersIndex++
		return task
	}
	return nil
}

func isInclude(args Args) bool {
	task, _ := args.Task()
	return task == "include"
}

func isHandlers(args Args) bool {
	task, _ := args.Task()
	return task == "handlers"
}
