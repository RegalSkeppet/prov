package prov

import (
	"errors"
	"os/exec"
	"path/filepath"
)

func init() {
	RegisterRunner("copy file", CopyFile)
}

func CopyFile(dir string, vars Vars, args Args, run bool) (Status, error) {
	dest, ok := args.String("destination")
	if !ok {
		return OK, ErrInvalidArg("destination")
	}
	if !filepath.IsAbs(dest) {
		return OK, errors.New(`argument "destination" needs to be absolute`)
	}
	src, ok := args.String("source")
	if !ok {
		return OK, ErrInvalidArg("source")
	}
	src = filepath.Join(dir, src)
	err := exec.Command("diff", src, dest).Run()
	status := OK
	if err != nil {
		if run {
			output, err := exec.Command("cp", src, dest).CombinedOutput()
			if err != nil {
				return OK, ErrCommandFailed{err, output}
			}
		}
		status = Changed
	}
	if run || status == OK {
		changed, err := SetFileProperties(dest, args, run)
		if err != nil {
			return OK, err
		}
		if changed {
			status = Changed
		}
	}
	return status, nil
}
