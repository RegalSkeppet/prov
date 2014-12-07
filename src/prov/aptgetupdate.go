package prov

import "os/exec"

func init() {
	RegisterRunner("apt-get update", AptGetUpdate)
}

var aptGetUpdateHasRun = false

func AptGetUpdate(dir string, vars, args map[interface{}]interface{}, live bool) (Status, error) {
	once, _ := getBoolVar(args, "once")
	if once && aptGetUpdateHasRun {
		return OK, nil
	}
	output, err := exec.Command("apt-get", "update").CombinedOutput()
	if err != nil {
		return OK, ErrCommandFailed{err, output}
	}
	aptGetUpdateHasRun = true
	return OK, nil
}
