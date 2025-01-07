package swclient

import (
	"path/filepath"

	hdconfig "github.com/nodeset-org/hyperdrive-daemon/shared/config"
)

var moduleApiKeyRelPath string = filepath.Join(hdconfig.SecretsDir, hdconfig.ModulesName)
