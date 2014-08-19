package prov

import (
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

func init() {
	RegisterRunner("template", Template)
}

func Template(dir string, vars Vars, args Args, run bool) (Status, error) {
	templateFile, ok := args.String("template")
	if !ok {
		return OK, ErrInvalidArg("template")
	}
	destination, ok := args.String("destination")
	if !ok {
		return OK, ErrInvalidArg("destination")
	}
	if !filepath.IsAbs(destination) {
		return OK, errors.New(`argument "destination" needs to be absolute`)
	}
	contents, err := ioutil.ReadFile(filepath.Join(dir, templateFile))
	if err != nil {
		return OK, err
	}
	templ, err := template.New("template").Parse(string(contents))
	if err != nil {
		return OK, err
	}
	file, err := ioutil.TempFile("", "")
	if err != nil {
		return OK, err
	}
	filename := file.Name()
	defer os.RemoveAll(filename)
	err = templ.Execute(file, vars)
	if err != nil {
		return OK, err
	}
	err = file.Close()
	if err != nil {
		return OK, err
	}
	status := OK
	err = exec.Command("diff", filename, destination).Run()
	if err != nil {
		if run {
			output, err := exec.Command("cp", filename, destination).CombinedOutput()
			if err != nil {
				return OK, ErrCommandFailed{err, output}
			}
		}
		status = Changed
	}
	if run || status == OK {
		changed, err := SetFileProperties(destination, args, run)
		if err != nil {
			return OK, err
		}
		if changed {
			status = Changed
		}
	}
	return status, nil
}
