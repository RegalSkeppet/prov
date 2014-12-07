package prov

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"os/exec"
	"regexp"
	"strings"
)

func init() {
	RegisterRunner("apt-key", AptKey)
}

var aptKeyGetFingerRE = regexp.MustCompile(`Key fingerprint = ([\d\w ]+)`)

func AptKey(dir string, vars, args map[interface{}]interface{}, live bool) (Status, error) {
	url, ok := getStringVar(args, "url")
	if !ok {
		return OK, ErrInvalidArg("url")
	}
	resp, err := http.Get(url)
	if err != nil {
		return OK, fmt.Errorf("apt-key could not fetch key: %s", err.Error())
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return OK, fmt.Errorf("apt-key could not fetch key: %s", resp.Status)
	}
	key, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return OK, fmt.Errorf("apt-key could not fetch key: %s", err.Error())
	}
	command := exec.Command("gpg", "--with-fingerprint")
	stdin, err := command.StdinPipe()
	if err != nil {
		return OK, fmt.Errorf("apt-key could not run gpg: %s", err.Error())
	}
	stdout, err := command.StdoutPipe()
	if err != nil {
		return OK, fmt.Errorf("apt-key could not run gpg: %s", err.Error())
	}
	err = command.Start()
	if err != nil {
		return OK, fmt.Errorf("apt-key could not run gpg: %s", err.Error())
	}
	_, err = stdin.Write(key)
	if err != nil {
		return OK, fmt.Errorf("apt-key could not run gpg: %s", err.Error())
	}
	err = stdin.Close()
	if err != nil {
		return OK, fmt.Errorf("apt-key could not run gpg: %s", err.Error())
	}
	output, err := ioutil.ReadAll(stdout)
	if err != nil {
		return OK, ErrCommandFailed{err, output}
	}
	err = command.Wait()
	if err != nil {
		return OK, ErrCommandFailed{err, output}
	}
	matches := aptKeyGetFingerRE.FindAllStringSubmatch(string(output), -1)
	if len(matches) == 0 || len(matches[0]) != 2 {
		return OK, fmt.Errorf("apt-key could not parse fingerprint: %s", output)
	}
	output, err = exec.Command("apt-key", "finger").CombinedOutput()
	if err != nil {
		return OK, ErrCommandFailed{err, output}
	}
	if strings.Contains(string(output), matches[0][1]) {
		return OK, nil
	}
	if live {
		command := exec.Command("apt-key", "add", "-")
		stdin, err = command.StdinPipe()
		if err != nil {
			return OK, fmt.Errorf("apt-key could not run apt-key: %s", err.Error())
		}
		stdout, err = command.StdoutPipe()
		if err != nil {
			return OK, fmt.Errorf("apt-key could not run apt-key: %s", err.Error())
		}
		stderr, err := command.StderrPipe()
		if err != nil {
			return OK, fmt.Errorf("apt-key could not run apt-key: %s", err.Error())
		}
		err = command.Start()
		if err != nil {
			return OK, fmt.Errorf("apt-key could not run apt-key: %s", err.Error())
		}
		_, err = stdin.Write(key)
		if err != nil {
			return OK, fmt.Errorf("apt-key could not run apt-key: %s", err.Error())
		}
		err = stdin.Close()
		if err != nil {
			return OK, fmt.Errorf("apt-key could not run apt-key: %s", err.Error())
		}
		buffer := bytes.NewBuffer(nil)
		_, err = io.Copy(buffer, stdout)
		if err != nil {
			return OK, fmt.Errorf("apt-key could not run apt-key: %s", err.Error())
		}
		_, err = io.Copy(buffer, stderr)
		if err != nil {
			return OK, fmt.Errorf("apt-key could not run apt-key: %s", err.Error())
		}
		err = command.Wait()
		if err != nil {
			return OK, ErrCommandFailed{err, buffer.Bytes()}
		}
	}
	return Changed, nil
}
