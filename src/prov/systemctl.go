package prov

import (
	"errors"
	"os/exec"
)

func init() {
	RegisterRunner("systemctl", Systemctl)
}

func Systemctl(dir string, vars, args map[interface{}]interface{}, live bool) (Status, error) {
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
		err := exec.Command("systemctl", "status", service).Run()
		if err == nil {
			return OK, nil
		}
		if live {
			output, err := exec.Command("systemctl", "start", service).CombinedOutput()
			if err != nil {
				return OK, ErrCommandFailed{err, output}
			}
		}
		return Changed, nil
	case "restarted":
		if live {
			output, err := exec.Command("systemctl", "restart", service).CombinedOutput()
			if err != nil {
				return OK, ErrCommandFailed{err, output}
			}
		}
		return Changed, nil
	default:
		return OK, errors.New(`unrecognized "state" variable`)
	}
}
