package prov

import (
	"errors"
	"log"
	"os/exec"
)

func init() {
	RegisterRunner("script", Script)
}

func Script(dir string, vars Vars, args Args, run bool) (Status, error) {
	ifArg, ok := args.String("if")
	if ok {
		err := exec.Command("bash", "-c", ifArg).Run()
		if err != nil {
			return OK, nil
		}
	}
	unless, ok := args.String("unless")
	if ok {
		err := exec.Command("bash", "-c", unless).Run()
		if err == nil {
			return OK, nil
		}
	}
	script, scriptOK := args.String("script")
	file, fileOK := args.String("file")
	if !scriptOK && !fileOK {
		return OK, errors.New(`need atleast one valid "script" or "file" argument`)
	}
	if scriptOK && fileOK {
		return OK, errors.New(`cannot use both "script" and "file" argument`)
	}
	if run {
		var output []byte
		var err error
		if scriptOK {
			output, err = exec.Command("bash", "-c", script).CombinedOutput()
		} else if fileOK {
			output, err = exec.Command("bash", file).CombinedOutput()
		} else {
			return OK, errors.New("neither script nor file was run")
		}
		log.Print(string(output))
		if err != nil {
			return OK, err
		}
	}
	return Changed, nil
}
