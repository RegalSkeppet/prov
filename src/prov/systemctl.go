package prov

import (
	"errors"
	"os/exec"
)

func init() {
	RegisterRunner("systemctl", Systemctl)
}

func Systemctl(dir string, vars, args map[interface{}]interface{}, live bool) (changed bool, err error) {
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
		err := exec.Command("systemctl", "status", service).Run()
		if err == nil {
			return false, nil
		}
		if live {
			output, err := exec.Command("systemctl", "start", service).CombinedOutput()
			if err != nil {
				return false, ErrCommandFailed{err, output}
			}
		}
		return true, nil
	case "restarted":
		if live {
			output, err := exec.Command("systemctl", "restart", service).CombinedOutput()
			if err != nil {
				return false, ErrCommandFailed{err, output}
			}
		}
		return true, nil
	default:
		return false, errors.New(`unrecognized "state" variable`)
	}
}
