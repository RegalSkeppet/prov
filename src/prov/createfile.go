package prov

import (
	"errors"
	"log"
	"os"
	"path/filepath"
)

func init() {
	RegisterRunner("create file", CreateFile)
}

func CreateFile(dir string, vars, args map[interface{}]interface{}, live bool) (changed bool, err error) {
	path, ok := getStringVar(args, "path")
	if !ok {
		return false, ErrInvalidArg("path")
	}
	if !filepath.IsAbs(path) {
		return false, errors.New(`argument "path" needs to be absolute`)
	}
	log.Printf("path is %q", path)
	file, err := os.Open(path)
	status := false
	if err == nil {
		log.Printf("file already exists")
		file.Close()
	} else {
		if live {
			file, err = os.Create(path)
			if err != nil {
				return false, err
			}
			file.Close()
		}
		log.Printf("file created")
		status = true
	}
	if live || status == false {
		changed, err := SetFileProperties(path, args, live)
		if err != nil {
			return false, err
		}
		if changed {
			log.Printf("file properties changed")
			status = true
		}
	}
	return status, nil
}
