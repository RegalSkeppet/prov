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

var templateFuncs = template.FuncMap{
	"smap": func(value interface{}) (map[string]interface{}, error) {
		valueMap, ok := value.(map[interface{}]interface{})
		if !ok {
			return nil, errors.New("value not a map")
		}
		result := make(map[string]interface{}, len(valueMap))
		for k, v := range valueMap {
			ks, ok := k.(string)
			if !ok {
				return nil, errors.New("key not a string")
			}
			result[ks] = v
		}
		return result, nil
	},
}

func Template(dir string, vars, args map[interface{}]interface{}, live bool) (Status, error) {
	templateFile, ok := getStringVar(args, "template")
	if !ok {
		return OK, ErrInvalidArg("template")
	}
	destination, ok := getStringVar(args, "destination")
	if !ok {
		return OK, ErrInvalidArg("destination")
	}
	if !filepath.IsAbs(destination) {
		return OK, errors.New(`argument "destination" needs to be absolute`)
	}
	extraVars, ok := getVarsVar(args, "vars")
	if ok {
		vars = copyVars(vars)
		setVars(vars, extraVars)
	}
	contents, err := ioutil.ReadFile(filepath.Join(dir, templateFile))
	if err != nil {
		return OK, err
	}
	templ, err := template.New("template").Funcs(templateFuncs).Parse(string(contents))
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
		if live {
			output, err := exec.Command("cp", filename, destination).CombinedOutput()
			if err != nil {
				return OK, ErrCommandFailed{err, output}
			}
		}
		status = Changed
	}
	if live || status == OK {
		changed, err := SetFileProperties(destination, args, live)
		if err != nil {
			return OK, err
		}
		if changed {
			status = Changed
		}
	}
	return status, nil
}
