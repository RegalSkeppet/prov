package prov

import (
	"errors"
	"os"
	"path/filepath"
)

func init() {
	RegisterRunner("create directory", CreateDir)
}

func CreateDir(dir string, vars, args map[interface{}]interface{}, live bool) (changed bool, err error) {
	path, ok := getStringVar(args, "path")
	if !ok {
		return false, ErrInvalidArg("path")
	}
	if !filepath.IsAbs(path) {
		return false, errors.New(`argument "path" needs to be absolute`)
	}
	file, err := os.Open(path)
	status := false
	if err == nil {
		file.Close()
	} else {
		if live {
			err = os.Mkdir(path, 0755)
			if err != nil {
				return false, err
			}
		}
		status = true
	}
	if live || status == false {
		changed, err := SetFileProperties(path, args, live)
		if err != nil {
			return false, err
		}
		if changed {
			status = true
		}
	}
	return status, nil
}
