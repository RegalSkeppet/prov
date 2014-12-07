package prov

import "os/exec"

func init() {
	RegisterRunner("pacman", Pacman)
}

func Pacman(dir string, vars, args map[interface{}]interface{}, live bool) (Status, error) {
	pack, ok := getStringVar(args, "package")
	if !ok {
		return OK, ErrInvalidArg("package")
	}
	err := exec.Command("pacman", "-Q", pack).Run()
	update, _ := getBoolVar(args, "update")
	if update || err != nil {
		output, err := exec.Command("pacman", "-Sy").CombinedOutput()
		if err != nil {
			return OK, ErrCommandFailed{err, output}
		}
	}
	if err == nil {
		if update {
			err := exec.Command("pacman", "-Qu", pack).Run()
			if err != nil {
				return OK, nil
			}
		} else {
			return OK, nil
		}
	}
	if live {
		output, err := exec.Command("pacman", "--noconfirm", "-S", pack).CombinedOutput()
		if err != nil {
			return OK, ErrCommandFailed{err, output}
		}
	}
	return Changed, nil
}
