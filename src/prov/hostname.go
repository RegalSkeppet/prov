package prov

import (
	"io/ioutil"
	"os"
	"os/exec"
)

func init() {
	RegisterRunner("hostname", Hostname)
}

func Hostname(dir string, vars, args map[interface{}]interface{}, live bool) (changed bool, err error) {
	hostname, ok := getStringVar(args, "hostname")
	if !ok {
		return false, ErrInvalidArg("hostname")
	}
	current, err := os.Hostname()
	if err != nil {
		return false, err
	}
	if current == hostname {
		return false, nil
	}
	if live {
		output, err := exec.Command("hostname", hostname).CombinedOutput()
		if err != nil {
			return false, ErrCommandFailed{err, output}
		}
		err = ioutil.WriteFile("/etc/hostname", []byte(hostname), 0644)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}
