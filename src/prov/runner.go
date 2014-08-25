package prov

import (
	"fmt"
	"log"
	"path/filepath"
	"time"
)

type Status string

const (
	OK      = Status("OK")
	Changed = Status("CHANGED")
)

type TaskRunner func(dir string, vars Vars, args Args, run bool) (Status, error)

var TaskRunners = map[string]TaskRunner{}

func RegisterRunner(name string, runner TaskRunner) {
	TaskRunners[name] = runner
}

func RunTask(dir string, vars Vars, args Args, run bool) (Status, error) {
	start := time.Now()
	name, ok := args.Name()
	if !ok {
		return OK, ErrTaskFailed{name, ErrInvalidArg("task")}
	}
	log.Printf("-- %s\n", name)
	task, ok := args.Task()
	if !ok {
		return OK, ErrTaskFailed{name, ErrInvalidArg("task")}
	}
	runner, ok := TaskRunners[task]
	if !ok {
		return OK, ErrTaskFailed{name, fmt.Errorf("unrecognized task: %q", task)}
	}
	status, err := runner(dir, vars, args, run)
	if err != nil {
		return OK, ErrTaskFailed{name, err}
	}
	log.Printf("%s (%s)\n", status, time.Since(start).String())
	return status, nil
}

func BootstrapFile(filename string, vars Vars, run bool) (ok, changed int, err error) {
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return 0, 0, err
	}
	return Include(
		filepath.Dir(absFilename),
		vars,
		Args{
			"task": "include",
			"path": filepath.Base(filename),
		},
		run,
	)
}
