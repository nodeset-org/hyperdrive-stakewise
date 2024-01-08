package hyperdrive

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/spf13/viper"
)

func TestSetViperAndWriteConfig(t *testing.T) {
	c := Config{
		DataDir:                "",
		Network:                "network",
		Name:                   "name",
		Vault:                  "vault",
		FeeRecipient:           "fee_recipient",
		ExceutionClientName:    "geth",
		ExceutionClientPort:    "30303",
		ExceutionClientAPIPort: "8545",
		ExceutionClientRPCPort: "8551",
		ConsensusClientName:    "nimbus",
		ConsensusClientPort:    "9000",
		ConsensusClientAPIPort: "5052",
		NumKeys:                "1",
	}
	c.SetViper()
	ecName := viper.GetString("ECNAME")
	if ecName != "geth" {
		t.Error("failed to set viper vars")

	}
	testDir := "TestSetViperAndWriteConfig"
	os.Mkdir(testDir, 0755)
	defer os.RemoveAll(testDir)
	SetConfigPath(testDir)
	err := c.WriteConfig()
	if err != nil {
		t.Errorf("failed to write config file: %v", err)
	}
	// Check that the output file contains the expected fields
	_, err = os.Stat(filepath.Join(testDir, "nodeset.env"))
	if err != nil {
		t.Errorf("output file does not exist: %v", err)
	}
	file, err := os.ReadFile(filepath.Join(testDir, "nodeset.env"))
	if err != nil {
		t.Errorf("failed to read output file: %v", err)
	}
	if !strings.Contains(string(file), "NETWORK") {
		t.Error("file does not contain NETWORK")
	}
	// Check that the output file contains the expected data
	config, err := LoadConfig()
	if err != nil {
		t.Errorf("viper failed to read config file: %v", err)
	}
	if config.ConsensusClientAPIPort != c.ConsensusClientAPIPort {
		t.Errorf("output file does not contain field")
	}

}
