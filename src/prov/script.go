package prov

import (
	"errors"
	"log"
	"os/exec"
)

func init() {
	RegisterRunner("script", Script)
}

func Script(dir string, vars, args map[interface{}]interface{}, live bool) (Status, error) {
	ifArg, ok := getStringVar(args, "if")
	if ok {
		err := exec.Command("bash", "-c", ifArg).Run()
		if err != nil {
			return OK, nil
		}
	}
	unless, ok := getStringVar(args, "unless")
	if ok {
		err := exec.Command("bash", "-c", unless).Run()
		if err == nil {
			return OK, nil
		}
	}
	script, scriptOK := getStringVar(args, "script")
	file, fileOK := getStringVar(args, "file")
	if !scriptOK && !fileOK {
		return OK, errors.New(`need atleast one valid "script" or "file" argument`)
	}
	if scriptOK && fileOK {
		return OK, errors.New(`cannot use both "script" and "file" argument`)
	}
	if live {
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
