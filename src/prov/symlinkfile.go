package prov

import (
	"bufio"
	"errors"
	"fmt"
	"log"
	"os"
	"os/exec"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
)

func init() {
	RegisterRunner("symlink file", SymlinkFile)
}

func SymlinkFile(dir string, vars, args map[interface{}]interface{}, live bool) (changed bool, err error) {
	dest, ok := getStringVar(args, "destination")
	if !ok {
		return false, ErrInvalidArg("destination")
	}
	if !filepath.IsAbs(dest) {
		return false, errors.New(`argument "destination" needs to be absolute`)
	}
	src, ok := getStringVar(args, "source")
	if !ok {
		return false, ErrInvalidArg("source")
	}
	src = filepath.Join(dir, src)
	log.Printf("looking for link %q -> %q", dest, src)
	existingSrc, err := os.Readlink(dest)
	if err != nil || existingSrc != src {
		if live {
			output, err := exec.Command("ln", "-fs", src, dest).CombinedOutput()
			if err != nil {
				return false, ErrCommandFailed{err, output}
			}
		}
		changed = true
	}
	if changed && !live {
		return true, nil
	}
	propsChanged, err := SetSymlinkProperties(dest, args, live)
	if propsChanged {
		changed = true
	}
	return
}

func SetSymlinkUser(filename, username string, live bool) (bool, error) {
	userData, err := user.Lookup(username)
	if err != nil {
		return false, err
	}
	uid, err := strconv.Atoi(userData.Uid)
	if err != nil {
		return false, err
	}
	var stat syscall.Stat_t
	err = syscall.Lstat(filename, &stat)
	if err != nil {
		return false, err
	}
	if int(stat.Uid) == uid {
		return false, nil
	}
	if live {
		err = os.Lchown(filename, uid, int(stat.Gid))
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func SetSymlinkGroup(filename, groupname string, live bool) (bool, error) {
	file, err := os.Open("/etc/group")
	if err != nil {
		return false, err
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	gid := -1
	for scanner.Scan() {
		line := strings.Split(scanner.Text(), ":")
		if line[0] == groupname {
			gid, err = strconv.Atoi(line[2])
			if err != nil {
				return false, err
			}
			break
		}
	}
	err = scanner.Err()
	if err != nil {
		return false, err
	}
	if gid < 0 {
		return false, fmt.Errorf("could not lookup group %q", groupname)
	}
	var stat syscall.Stat_t
	err = syscall.Lstat(filename, &stat)
	if err != nil {
		return false, err
	}
	if int(stat.Gid) == gid {
		return false, nil
	}
	if live {
		err = os.Lchown(filename, int(stat.Uid), gid)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func SetSymlinkProperties(filename string, args map[interface{}]interface{}, live bool) (bool, error) {
	var result bool
	owner, ok := getStringVar(args, "owner")
	if ok {
		changed, err := SetSymlinkUser(filename, owner, live)
		if err != nil {
			return false, err
		}
		if changed {
			result = true
			log.Printf("changed owner to %q", owner)
		} else {
			log.Printf("owner already %q", owner)
		}
	}
	group, ok := getStringVar(args, "group")
	if ok {
		changed, err := SetSymlinkGroup(filename, group, live)
		if err != nil {
			return false, err
		}
		if changed {
			result = true
			log.Printf("changed group to %q", group)
		} else {
			log.Printf("group already %q", group)
		}
	}
	return result, nil
}
