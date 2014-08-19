package prov

import (
	"bufio"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

func init() {
	RegisterRunner("user", User)
}

var uidRE = regexp.MustCompile(`(^|\s)uid=(\d+)`)
var gidRE = regexp.MustCompile(`(^|\s)gid=(\d+)`)

func User(dir string, vars Vars, args Args, run bool) (Status, error) {
	username, ok := args.String("user")
	if !ok {
		return OK, ErrInvalidArg("user")
	}
	status := OK
	err := exec.Command("id", username).Run()
	if err != nil {
		output, err := exec.Command("useradd", "--create-home", username).CombinedOutput()
		if err != nil {
			return OK, ErrCommandFailed{err, output}
		}
		status = Changed
	}
	keys, ok := args.StringList("keys")
	if ok {
		userInfo, err := user.Lookup(username)
		if err != nil {
			return OK, err
		}
		uid, err := strconv.Atoi(userInfo.Uid)
		if err != nil {
			return OK, err
		}
		gid, err := strconv.Atoi(userInfo.Gid)
		if err != nil {
			return OK, err
		}
		sshDir := filepath.Join("/home", username, ".ssh")
		authKeyFilename := filepath.Join(sshDir, "authorized_keys")
		_, err = os.Stat(sshDir)
		if err != nil {
			err = os.Mkdir(sshDir, 0700)
			if err != nil {
				return OK, err
			}
			err = os.Chown(sshDir, uid, gid)
			if err != nil {
				return OK, err
			}
			status = Changed
		}
		_, err = os.Stat(authKeyFilename)
		chown := err != nil
		file, err := os.OpenFile(authKeyFilename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
		if err != nil {
			return OK, err
		}
		if chown {
			err = file.Chown(uid, gid)
			if err != nil {
				return OK, err
			}
		}
		defer file.Close()
		var oldKeys []string
		scanner := bufio.NewScanner(file)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}
			oldKeys = append(oldKeys, line)
		}
		err = scanner.Err()
		if err != nil {
			return OK, err
		}
		dirty := false
		if len(keys) != len(oldKeys) {
			dirty = true
		} else {
			sort.Strings(keys)
			sort.Strings(oldKeys)
			for i := range keys {
				if keys[i] != oldKeys[i] {
					dirty = true
					break
				}
			}
		}
		if dirty {
			err = file.Truncate(0)
			if err != nil {
				return OK, err
			}
			_, err = file.Seek(0, 0)
			if err != nil {
				return OK, err
			}
			_, err = file.Write([]byte(strings.Join(keys, "\n")))
			if err != nil {
				return OK, err
			}
			status = Changed
		}
	}
	return status, nil
}
