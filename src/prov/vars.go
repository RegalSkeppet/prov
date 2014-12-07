package prov

func getBoolVar(vars map[interface{}]interface{}, name string) (bool, bool) {
	value, ok := vars[name]
	if !ok {
		return false, false
	}
	result, ok := value.(bool)
	return result, ok
}

func getIntVar(vars map[interface{}]interface{}, name string) (int, bool) {
	value, ok := vars[name]
	if !ok {
		return 0, false
	}
	result, ok := value.(int)
	return result, ok
}

func getStringVar(vars map[interface{}]interface{}, name string) (string, bool) {
	value, ok := vars[name]
	if !ok {
		return "", false
	}
	result, ok := value.(string)
	return result, ok
}

func getStringListVar(vars map[interface{}]interface{}, name string) ([]string, bool) {
	value1, ok := vars[name]
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

func getVarsVar(vars map[interface{}]interface{}, name string) (map[interface{}]interface{}, bool) {
	value1, ok := vars[name]
	if !ok {
		return nil, false
	}
	value2, ok := value1.(map[interface{}]interface{})
	if !ok {
		return nil, false
	}
	return value2, true
}

func copyVars(vars map[interface{}]interface{}) map[interface{}]interface{} {
	other := make(map[interface{}]interface{}, len(vars))
	for k, v := range vars {
		other[k] = v
	}
	return other
}

func setVars(upstream, downstream map[interface{}]interface{}) {
	for k, v := range downstream {
		_, ok := upstream[k]
		if !ok {
			upstream[k] = v
		}
	}
}
