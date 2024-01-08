package hyperdrive

import (
	"os"
	"os/user"
	"testing"
)

func TestIsRoot(t *testing.T) {
	// Test IsRoot with different UIDs and usernames
	u, err := user.Current()
	if u.Name == "root" {
		t.Errorf("Current User should not test with root: %v", err)
	}
	if IsRoot() {
		t.Log("Current user is root")
	} else {
		t.Log("Current user is not root")
	}
}

// func TestCallingUser(t *testing.T) {
// 	// Test CallingUser with different environments
// 	u, err := user.Current()
// 	if err != nil {
// 		t.Errorf("Failed to get current user: %v", err)
// 		return
// 	}
// 	callingUser := os.Getenv("SUDO_USER")
// 	if callingUser == "" {
// 		t.Log("Calling user is empty")
// 		return
// 	}
// 	u2, err := user.Lookup(callingUser)
// 	if err != nil {
// 		t.Errorf("Failed to lookup calling user: %v", err)
// 		return
// 	}
// 	if !u.Equal(u2) {
// 		t.Log("Calling user is not equal to looked up user")
// 		return
// 	}
// }

func TestCallingUserHomeDir(t *testing.T) {
	// Test CallingUserHomeDir with different home directories
	u, err := user.Current()
	if err != nil {
		t.Errorf("Failed to get current user: %v", err)
		return
	}
	homeDir := ""
	if IsRoot() {
		homeDir = os.Getenv("SUDO_HOME")
	} else {
		homeDir = u.HomeDir
	}
	if homeDir == "" {
		t.Log("Calling user home directory is empty")
		return
	}
	err = Chown(homeDir, u)
	if err != nil {
		t.Errorf("Failed to chown calling user home directory: %v", err)
		return
	}
}

func TestChown(t *testing.T) {
	// Test Chown with different directories and users
	dir := "testdir"

	u, err := user.Lookup("nobody")
	if err != nil {
		t.Errorf("Failed to get current user: %v", err)
		return
	}
	err = Chown(dir, u)
	if err == nil {
		t.Logf("Chowned %v with UID %v successfully", dir, u.Uid)
		return
	}

}
