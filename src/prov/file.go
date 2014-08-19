package prov

import (
	"bufio"
	"fmt"
	"os"
	"os/user"
	"strconv"
	"strings"
	"syscall"
)

func SetFileMode(filename string, mode os.FileMode, run bool) (bool, error) {
	stat, err := os.Stat(filename)
	if err != nil {
		return false, err
	}
	if stat.Mode().Perm() == mode {
		return false, nil
	}
	if run {
		err = os.Chmod(filename, mode)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func SetFileUser(filename, username string, run bool) (bool, error) {
	userData, err := user.Lookup(username)
	if err != nil {
		return false, err
	}
	uid, err := strconv.Atoi(userData.Uid)
	if err != nil {
		return false, err
	}
	var stat syscall.Stat_t
	err = syscall.Stat(filename, &stat)
	if err != nil {
		return false, err
	}
	if int(stat.Uid) == uid {
		return false, nil
	}
	if run {
		err = os.Chown(filename, uid, int(stat.Gid))
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func SetFileGroup(filename, groupname string, run bool) (bool, error) {
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
	err = syscall.Stat(filename, &stat)
	if err != nil {
		return false, err
	}
	if int(stat.Gid) == gid {
		return false, nil
	}
	if run {
		err = os.Chown(filename, int(stat.Uid), gid)
		if err != nil {
			return false, err
		}
	}
	return true, nil
}

func SetFileProperties(filename string, args Args, run bool) (bool, error) {
	var result bool
	mode, ok := args.Int("mode")
	if ok {
		changed, err := SetFileMode(filename, os.FileMode(mode), run)
		if err != nil {
			return false, err
		}
		if changed {
			result = true
		}
	}
	owner, ok := args.String("owner")
	if ok {
		changed, err := SetFileUser(filename, owner, run)
		if err != nil {
			return false, err
		}
		if changed {
			result = true
		}
	}
	group, ok := args.String("group")
	if ok {
		changed, err := SetFileGroup(filename, group, run)
		if err != nil {
			return false, err
		}
		if changed {
			result = true
		}
	}
	return result, nil
}
