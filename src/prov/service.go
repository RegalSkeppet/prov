package prov

import (
	"errors"
	"os/exec"
	"strings"
)

func init() {
	RegisterRunner("service", Service)
}

func Service(dir string, vars Vars, args Args, run bool) (Status, error) {
	service, ok := args.String("service")
	if !ok {
		return OK, ErrInvalidArg("service")
	}
	state, ok := args.String("state")
	if !ok {
		return OK, ErrInvalidArg("state")
	}
	switch state {
	case "started":
		output, err := exec.Command("service", service, "status").CombinedOutput()
		if err != nil {
			return OK, ErrCommandFailed{err, output}
		}
		if strings.Contains(string(output), "start/running") {
			return OK, nil
		}
		if run {
			output, err := exec.Command("service", service, "start").CombinedOutput()
			if err != nil {
				return OK, ErrCommandFailed{err, output}
			}
		}
		return Changed, nil
	case "restarted":
		if run {
			output, err := exec.Command("service", service, "restart").CombinedOutput()
			if err != nil {
				return OK, ErrCommandFailed{err, output}
			}
		}
		return Changed, nil
	default:
		return OK, errors.New(`unrecognized "state" variable`)
	}
}
