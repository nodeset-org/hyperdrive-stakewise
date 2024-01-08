package hyperdrive

import (
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/spf13/viper"
)

type Command struct {
	Text string
	Cmd  *exec.Cmd
}

func (c Config) ExecCommand(text string) error {
	cmd, err := c.BuildCommand(text)
	if err != nil {
		return err
	}

	// Start the command
	if err := cmd.Start(); err != nil {
		return err
	}

	return cmd.Wait()
}

type StringSlice interface {
	string | []string
}

// Builds a command incuding setting the dataDir and environ from Config and os.Environ()
func (c Config) BuildCommand(text string) (*exec.Cmd, error) {

	// var cmd *exec.Cmd
	// switch text.(type) {
	// case string:
	// 	txt := text.(string)
	// 	cmd = exec.Command("sh", "-c", txt)
	// case []string:
	// 	txt := text.([]string)
	// 	if len(txt) == 1 {
	// 		cmd = exec.Command(txt[0])
	// 	} else {
	// 		cmd = exec.Command(txt[0], txt[1:]...)
	// 	}

	// }

	cmd := exec.Command("sh", "-c", text)

	cmd.Dir = c.DataDir
	cmd.Env = append(cmd.Env, os.Environ()...)
	cmd.Env = append(cmd.Env, fmt.Sprintf("PATH=%s", os.Getenv("PATH")))
	cmd.Env = append(cmd.Env, fmt.Sprintf("DATA_DIR=%s", c.DataDir))

	for k, v := range viper.AllSettings() {
		env := fmt.Sprintf("%s=%s", strings.ToUpper(k), v)
		cmd.Env = append(cmd.Env, env)
	}

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd, nil

}
