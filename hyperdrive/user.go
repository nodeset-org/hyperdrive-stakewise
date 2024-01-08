package hyperdrive

import (
	"os"
	"os/user"
	"runtime"
	"strconv"
	"syscall"
)

func IsRoot() bool {
	u, _ := user.Current()
	if os.Geteuid() == 0 && u.Username == "root" {
		return true
	} else {
		return false
	}
}

func CallingUser() (*user.User, error) {

	u, err := user.Current()
	if err != nil {
		return u, err
	}
	if IsRoot() {
		callingUser := os.Getenv("SUDO_USER")
		u, err = user.Lookup(callingUser)
		if err != nil {
			return u, err
		}
	}
	return u, nil

}

func CallingUserHomeDir() (string, error) {
	u, err := CallingUser()
	return u.HomeDir, err
}

func Chown(dir string, u *user.User) error {

	if runtime.GOOS != "windows" {
		uid, _ := strconv.Atoi(u.Uid)
		gid, _ := strconv.Atoi(u.Gid)

		err := syscall.Chown(dir, uid, gid)
		if err != nil {
			return err
		}
	}
	return nil
}
