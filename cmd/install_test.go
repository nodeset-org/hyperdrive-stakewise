package cmd

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func Test_InstallCommand(t *testing.T) {

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	testDir := "Test_InstallCommand"
	os.Mkdir(testDir, 0755)
	defer os.RemoveAll(testDir)
	rootCmd.SetArgs([]string{
		"init",
		"--network=holskey",
		fmt.Sprintf("--directory=./%s/", testDir),
		"--ecname=geth",
	})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}
	rootCmd.SetArgs([]string{
		"install",
		fmt.Sprintf("--directory=./%s/", testDir),
	})
	err = rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

}
