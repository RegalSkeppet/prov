package prov

import "os/exec"

func init() {
	RegisterRunner("apt-get update", AptGetUpdate)
}

var aptGetUpdateHasRun = false

func AptGetUpdate(dir string, vars Vars, args Args, run bool) (Status, error) {
	once := args.Bool("once")
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
