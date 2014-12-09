package prov

import "os/exec"

func init() {
	RegisterRunner("apt-get update", AptGetUpdate)
}

var aptGetUpdateHasRun = false

func AptGetUpdate(dir string, vars, args map[interface{}]interface{}, live bool) (changed bool, err error) {
	once, _ := getBoolVar(args, "once")
	if once && aptGetUpdateHasRun {
		return false, nil
	}
	output, err := exec.Command("apt-get", "update").CombinedOutput()
	if err != nil {
		return false, ErrCommandFailed{err, output}
	}
	aptGetUpdateHasRun = true
	return false, nil
}
