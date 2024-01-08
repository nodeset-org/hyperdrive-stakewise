package hyperdrive

import (
	"errors"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/spf13/viper"
)

var (
	ConfigFile                = "nodeset.env"
	ErrorCanNotFindConfigFile = errors.New("Are you sure this data directory is correct? If so, you must recover your configuration manually. Try running the init command")
)

type Config struct {
	DataDir                string `mapstructure:"DATA_DIR"`
	Network                string `mapstructure:"NETWORK"`
	Name                   string `mapstructure:"NAME"`
	Vault                  string `mapstructure:"VAULT"`
	FeeRecipient           string `mapstructure:"FEERECIPIENT"`
	ExceutionClientName    string `mapstructure:"ECNAME"`
	ExceutionClientPort    string `mapstructure:"ECPORT"`
	ExceutionClientAPIPort string `mapstructure:"ECAPIPORT"`
	ExceutionClientRPCPort string `mapstructure:"ECRPCPORT"`
	// http://$ECNAME
	ExceutionClientURL     string `mapstructure:"ECURL"`
	ConsensusClientName    string `mapstructure:"CCNAME"`
	ConsensusClientPort    string `mapstructure:"CCPORT"`
	ConsensusClientAPIPort string `mapstructure:"CCAPIPORT"`
	//CCURL=http://$CCNAME
	ConsensusClientURL string `mapstructure:"CCURL"`
	NumKeys            string `mapstructure:"NUMKEYS"`
	Checkpoint         bool   `mapstructure:"CHECKPOINT"`
	CheckpointURL      string `mapstructure:"CHECKPOINT_URL"`
	InternalClients    bool   `mapstructure:"INTERNAL_CLIENTS"`
}

var (
	Holskey = Config{
		Network:                "holesky",
		Name:                   "holesky",
		Vault:                  "0x646F5285D195e08E309cF9A5aDFDF68D6Fcc51C4",
		FeeRecipient:           "0xc98F25BcAA6B812a07460f18da77AF8385be7b56",
		ExceutionClientName:    "",
		ExceutionClientPort:    "30303",
		ExceutionClientAPIPort: "8545",
		ExceutionClientRPCPort: "8551",
		ConsensusClientName:    "nimbus",
		ConsensusClientPort:    "9000",
		ConsensusClientAPIPort: "5052",
		NumKeys:                "1",
	}
	HoleskyDev = Config{
		Network:                "holesky",
		Name:                   "holesky",
		Vault:                  "0x01b353Abc66A65c4c0Ac9c2ecF82e693Ce0303Bc",
		FeeRecipient:           "0xc98F25BcAA6B812a07460f18da77AF8385be7b56",
		ExceutionClientName:    "",
		ExceutionClientPort:    "30303",
		ExceutionClientAPIPort: "8545",
		ExceutionClientRPCPort: "8551",
		ConsensusClientName:    "",
		ConsensusClientPort:    "9000",
		ConsensusClientAPIPort: "5052",
		NumKeys:                "1",
	}

	Gravita = Config{
		Network:                "mainnet",
		Name:                   "gravita",
		Vault:                  "",
		FeeRecipient:           "",
		ExceutionClientName:    "",
		ExceutionClientPort:    "30303",
		ExceutionClientAPIPort: "8545",
		ExceutionClientRPCPort: "8551",
		ConsensusClientName:    "",
		ConsensusClientPort:    "9000",
		ConsensusClientAPIPort: "5052",
		NumKeys:                "1",
	}
)

func (c Config) SetViper() {
	// Use reflect to get the type and value of the  struct
	t := reflect.TypeOf(c)
	v := reflect.ValueOf(c)

	// Loop over each field in the struct, set the viper keys for the mapstructure tag
	for i := 0; i < v.NumField(); i++ {
		field := t.Field(i)
		k := field.Tag.Get("mapstructure")
		value := v.Field(i).Interface()
		if value != "" {
			viper.Set(k, v.Field(i).Interface())
		}

	}

}

func SetConfigPath(dir string) error {
	dirAbs, err := filepath.Abs(dir)
	if err != nil {
		return err
	}
	viper.Set("DATA_DIR", dirAbs)
	viper.AddConfigPath(dirAbs)

	parts := strings.Split(ConfigFile, ".")
	name := parts[0]
	ty := parts[1]
	viper.SetConfigName(name)
	viper.SetConfigType(ty)
	return nil
}

func (c *Config) WriteConfig() error {

	err := viper.SafeWriteConfig()
	if err != nil {
		var alreadyExistsErr viper.ConfigFileAlreadyExistsError
		if errors.Is(err, &alreadyExistsErr) {
			viper.WriteConfig()
		}
	}
	return err
}

func LoadConfig() (Config, error) {
	var c Config
	if err := viper.ReadInConfig(); err != nil {
		return c, err
	}
	if err := viper.Unmarshal(&c); err != nil {
		return c, err
	}
	return c, nil
}
