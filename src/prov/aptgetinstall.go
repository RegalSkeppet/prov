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

func AptGetInstall(dir string, vars Vars, args Args, run bool) (Status, error) {
	pack, ok := args.String("package")
	if !ok {
		return OK, ErrInvalidArg("package")
	}
	params := []string{"--simulate", "--yes", "install", pack}
	if run {
		params = params[1:]
	}
	output, err := exec.Command("apt-get", params...).CombinedOutput()
	if err != nil {
		return OK, ErrCommandFailed{err, output}
	}
	return aptGetInstallHasChanged(output)
}

func aptGetInstallHasChanged(output []byte) (Status, error) {
	matches := aptGetInstallRE.FindAllStringSubmatch(string(output), -1)
	if len(matches) == 0 {
		return OK, fmt.Errorf("apt-get install could not match output: %s", output)
	}
	if len(matches[0]) != 5 {
		return OK, fmt.Errorf("apt-get install could not match output: %s", output)
	}
	upgraded, err := strconv.Atoi(matches[0][1])
	if err != nil {
		return OK, fmt.Errorf("apt-get install could not match output: %s", err.Error())
	}
	installed, err := strconv.Atoi(matches[0][2])
	if err != nil {
		return OK, fmt.Errorf("apt-get install could not match output: %s", err.Error())
	}
	if upgraded+installed > 0 {
		return Changed, nil
	}
	return OK, nil
}
