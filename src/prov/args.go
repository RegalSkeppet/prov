package prov

import "fmt"

type ErrInvalidArg string

func (me ErrInvalidArg) Error() string {
	return fmt.Sprintf("missing or invalid argument: %q", string(me))
}

type ErrCommandFailed struct {
	Err    error
	Output []byte
}

func (me ErrCommandFailed) Error() string {
	return fmt.Sprintf("%s: %s", me.Err.Error(), me.Output)
}

type ErrTaskFailed struct {
	Name string
	Err  error
}

func (me ErrTaskFailed) Error() string {
	return fmt.Sprintf("task %q failed: %s", me.Name, me.Err.Error())
}

type Args map[string]interface{}

func (me Args) Name() (string, bool) {
	name, ok := me.String("name")
	if !ok {
		return me.Task()
	}
	return name, ok
}

func (me Args) Task() (string, bool) {
	return me.String("task")
}

func (me Args) Bool(name string) bool {
	value, ok := me[name]
	if !ok {
		return false
	}
	result, ok := value.(bool)
	return result && ok
}

func (me Args) Int(name string) (int, bool) {
	value, ok := me[name]
	if !ok {
		return 0, false
	}
	result, ok := value.(int)
	return result, ok
}

func (me Args) String(name string) (string, bool) {
	value, ok := me[name]
	if !ok {
		return "", false
	}
	result, ok := value.(string)
	return result, ok
}

func (me Args) StringList(name string) ([]string, bool) {
	value1, ok := me[name]
	if !ok {
		return nil, false
	}
	value2, ok := value1.([]interface{})
	if !ok {
		return nil, false
	}
	result := make([]string, len(value2))
	for i, value3 := range value2 {
		value4, ok := value3.(string)
		if !ok {
			return nil, false
		}
		result[i] = value4
	}
	return result, true
}

func (me Args) Vars(name string) (Vars, bool) {
	value1, ok := me[name]
	if !ok {
		return nil, false
	}
	value2, ok := value1.(map[interface{}]interface{})
	if !ok {
		return nil, false
	}
	result := make(Vars, len(value2))
	for kI, v := range value2 {
		k, ok := kI.(string)
		if !ok {
			return nil, false
		}
		result[k] = v
	}
	return result, true
}
