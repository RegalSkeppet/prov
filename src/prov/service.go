package prov

import (
	"errors"
	"os/exec"
	"strings"
)

func init() {
	RegisterRunner("service", Service)
}

func Service(dir string, vars, args map[interface{}]interface{}, live bool) (Status, error) {
	service, ok := getStringVar(args, "service")
	if !ok {
		return OK, ErrInvalidArg("service")
	}
	state, ok := getStringVar(args, "state")
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
		if live {
			output, err := exec.Command("service", service, "start").CombinedOutput()
			if err != nil {
				return OK, ErrCommandFailed{err, output}
			}
		}
		return Changed, nil
	case "restarted":
		if live {
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
