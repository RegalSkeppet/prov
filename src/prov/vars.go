package prov

import "gopkg.in/yaml.v1"

type Vars map[string]interface{}

func (vars *Vars) String() string {
	return ""
}

func (vars *Vars) Set(raw string) error {
	return yaml.Unmarshal([]byte(raw), vars)
}

func (me Vars) SetVars(vars Vars) {
	for name, value := range vars {
		_, ok := me[name]
		if !ok {
			me[name] = value
		}
	}
}

func (me Vars) GetString(name string) (string, bool) {
	value, ok := me[name]
	if !ok {
		return "", false
	}
	result, ok := value.(string)
	return result, ok
}
