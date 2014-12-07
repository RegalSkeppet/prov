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

type TaskRunner func(dir string, vars, args map[interface{}]interface{}, live bool) (Status, error)

var TaskRunners = map[string]TaskRunner{}

func RegisterRunner(name string, runner TaskRunner) {
	TaskRunners[name] = runner
}

func RunTask(dir string, key interface{}, vars, args map[interface{}]interface{}, live bool) (Status, error) {
	start := time.Now()
	log.Printf("-- %v\n", key)
	task, ok := getStringVar(args, "task")
	if !ok {
		return OK, ErrTaskFailed{key, ErrInvalidArg("task")}
	}
	runner, ok := TaskRunners[task]
	if !ok {
		return OK, ErrTaskFailed{key, fmt.Errorf("unrecognized task: %q", task)}
	}
	status, err := runner(dir, vars, args, live)
	if err != nil {
		return OK, ErrTaskFailed{key, err}
	}
	log.Printf("%s (%s)\n", status, time.Since(start).String())
	return status, nil
}

func BootstrapFile(filename string, vars map[interface{}]interface{}, live bool) (ok, changed int, err error) {
	absFilename, err := filepath.Abs(filename)
	if err != nil {
		return 0, 0, err
	}
	return Include(
		filepath.Dir(absFilename),
		vars,
		map[interface{}]interface{}{
			"task": "include",
			"path": filepath.Base(filename),
		},
		live,
	)
}
