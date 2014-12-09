package prov

import (
	"errors"
	"os/exec"
	"strings"
)

func init() {
	RegisterRunner("service", Service)
}

func Service(dir string, vars, args map[interface{}]interface{}, live bool) (changed bool, err error) {
	service, ok := getStringVar(args, "service")
	if !ok {
		return false, ErrInvalidArg("service")
	}
	state, ok := getStringVar(args, "state")
	if !ok {
		return false, ErrInvalidArg("state")
	}
	switch state {
	case "started":
		output, err := exec.Command("service", service, "status").CombinedOutput()
		if err != nil {
			return false, ErrCommandFailed{err, output}
		}
		if strings.Contains(string(output), "start/running") {
			return false, nil
		}
		if live {
			output, err := exec.Command("service", service, "start").CombinedOutput()
			if err != nil {
				return false, ErrCommandFailed{err, output}
			}
		}
		return true, nil
	case "restarted":
		if live {
			output, err := exec.Command("service", service, "restart").CombinedOutput()
			if err != nil {
				return false, ErrCommandFailed{err, output}
			}
		}
		return true, nil
	default:
		return false, errors.New(`unrecognized "state" variable`)
	}
}
