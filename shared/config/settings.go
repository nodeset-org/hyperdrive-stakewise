package swconfig

const (
	ModuleName           string = "stakewise"
	ShortModuleName      string = "sw"
	DaemonBaseRoute      string = ModuleName
	ApiVersion           string = "1"
	ApiClientRoute       string = DaemonBaseRoute + "/api/v" + ApiVersion
	WalletFilename       string = "wallet.json"
	PasswordFilename     string = "password.txt"
	KeystorePasswordFile string = "secret.txt"
	DepositDataFile      string = "deposit-data.json"
	AvailableKeysFile    string = "available-keys.json"
	OracleManagerFile    string = "oracle-data.json"
	DefaultApiPort       uint16 = 8180
	RelayLogName         string = "relay.log"
	DefaultRelayPort     uint16 = 18180

	// Volumes
	DataVolume string = "swdata"

	// Logging
	ClientLogName string = "hd.log"
)
