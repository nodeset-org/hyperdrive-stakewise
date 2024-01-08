package cmd

import (
	"bytes"
	"fmt"
	"os"
	"testing"
)

func Test_InitCommand(t *testing.T) {

	b := bytes.NewBufferString("")
	rootCmd.SetOut(b)
	testDir := "Test_InitCommand"
	os.Mkdir(testDir, 0755)
	defer os.RemoveAll(testDir)
	rootCmd.SetArgs([]string{
		"init",
		"--network=holskey-dev",
		fmt.Sprintf("--directory=./%s/", testDir),
		"--ecname=nethermind",
		"--config=nodeset.json",
	})
	err := rootCmd.Execute()
	if err != nil {
		t.Fatal(err)
	}

}
