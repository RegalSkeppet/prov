package prov

import (
	"errors"
	"os/exec"
	"path/filepath"
)

func init() {
	RegisterRunner("copy file", CopyFile)
}

func CopyFile(dir string, vars, args map[interface{}]interface{}, live bool) (changed bool, err error) {
	dest, ok := getStringVar(args, "destination")
	if !ok {
		return false, ErrInvalidArg("destination")
	}
	if !filepath.IsAbs(dest) {
		return false, errors.New(`argument "destination" needs to be absolute`)
	}
	src, ok := getStringVar(args, "source")
	if !ok {
		return false, ErrInvalidArg("source")
	}
	src = filepath.Join(dir, src)
	err = exec.Command("diff", src, dest).Run()
	status := false
	if err != nil {
		if live {
			output, err := exec.Command("cp", src, dest).CombinedOutput()
			if err != nil {
				return false, ErrCommandFailed{err, output}
			}
		}
		status = true
	}
	if live || status == false {
		changed, err := SetFileProperties(dest, args, live)
		if err != nil {
			return false, err
		}
		if changed {
			status = true
		}
	}
	return status, nil
}
