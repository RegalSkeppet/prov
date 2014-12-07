package prov

import (
	"errors"
	"os"
	"path/filepath"
)

func init() {
	RegisterRunner("create file", CreateFile)
}

func CreateFile(dir string, vars, args map[interface{}]interface{}, live bool) (Status, error) {
	path, ok := getStringVar(args, "path")
	if !ok {
		return OK, ErrInvalidArg("path")
	}
	if !filepath.IsAbs(path) {
		return OK, errors.New(`argument "path" needs to be absolute`)
	}
	file, err := os.Open(path)
	status := OK
	if err == nil {
		file.Close()
	} else {
		if live {
			file, err = os.Create(path)
			if err != nil {
				return OK, err
			}
			file.Close()
		}
		status = Changed
	}
	if live || status == OK {
		changed, err := SetFileProperties(path, args, live)
		if err != nil {
			return OK, err
		}
		if changed {
			status = Changed
		}
	}
	return status, nil
}
