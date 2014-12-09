package prov

import (
	"fmt"
	"os/exec"
	"regexp"
	"strconv"
)

func init() {
	RegisterRunner("apt-get install", AptGetInstall)
}

var aptGetInstallRE = regexp.MustCompile(`(?m)^(\d+) upgraded, (\d+) newly installed, (\d+) to remove and (\d+) not upgraded\.`)

func AptGetInstall(dir string, vars, args map[interface{}]interface{}, live bool) (changed bool, err error) {
	pack, ok := getStringVar(args, "package")
	if !ok {
		return false, ErrInvalidArg("package")
	}
	params := []string{"--simulate", "--yes", "install", pack}
	if live {
		params = params[1:]
	}
	output, err := exec.Command("apt-get", params...).CombinedOutput()
	if err != nil {
		return false, ErrCommandFailed{err, output}
	}
	return aptGetInstallHasChanged(output)
}

func aptGetInstallHasChanged(output []byte) (changed bool, err error) {
	matches := aptGetInstallRE.FindAllStringSubmatch(string(output), -1)
	if len(matches) == 0 {
		return false, fmt.Errorf("apt-get install could not match output: %s", output)
	}
	if len(matches[0]) != 5 {
		return false, fmt.Errorf("apt-get install could not match output: %s", output)
	}
	upgraded, err := strconv.Atoi(matches[0][1])
	if err != nil {
		return false, fmt.Errorf("apt-get install could not match output: %s", err.Error())
	}
	installed, err := strconv.Atoi(matches[0][2])
	if err != nil {
		return false, fmt.Errorf("apt-get install could not match output: %s", err.Error())
	}
	if upgraded+installed > 0 {
		return true, nil
	}
	return false, nil
}
