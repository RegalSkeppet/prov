package prov

import (
	"errors"
	"os/exec"
	"path/filepath"
)

func init() {
	RegisterRunner("copy file", CopyFile)
}

func CopyFile(dir string, vars, args map[interface{}]interface{}, live bool) (Status, error) {
	dest, ok := getStringVar(args, "destination")
	if !ok {
		return OK, ErrInvalidArg("destination")
	}
	if !filepath.IsAbs(dest) {
		return OK, errors.New(`argument "destination" needs to be absolute`)
	}
	src, ok := getStringVar(args, "source")
	if !ok {
		return OK, ErrInvalidArg("source")
	}
	src = filepath.Join(dir, src)
	err := exec.Command("diff", src, dest).Run()
	status := OK
	if err != nil {
		if live {
			output, err := exec.Command("cp", src, dest).CombinedOutput()
			if err != nil {
				return OK, ErrCommandFailed{err, output}
			}
		}
		status = Changed
	}
	if live || status == OK {
		changed, err := SetFileProperties(dest, args, live)
		if err != nil {
			return OK, err
		}
		if changed {
			status = Changed
		}
	}
	return status, nil
}
