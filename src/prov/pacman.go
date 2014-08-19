package prov

import "os/exec"

func init() {
	RegisterRunner("pacman", Pacman)
}

func Pacman(dir string, vars Vars, args Args, run bool) (Status, error) {
	pack, ok := args.String("package")
	if !ok {
		return OK, ErrInvalidArg("package")
	}
	err := exec.Command("pacman", "-Q", pack).Run()
	update := args.Bool("update")
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
	if run {
		output, err := exec.Command("pacman", "--noconfirm", "-S", pack).CombinedOutput()
		if err != nil {
			return OK, ErrCommandFailed{err, output}
		}
	}
	return Changed, nil
}
