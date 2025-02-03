package utils

const (
	// Name of the log file for the adapter
	AdapterLogFile string = "adapter.log"

	// Service log file
	ServiceLogFile string = "service.log"

	// Adapter configuration file
	AdapterConfigFile string = "adapter-cfg.yaml"

	// Service configuration file
	ServiceConfigFile string = "service-cfg.yaml"

	// Module name as it appears in the descriptor
	ModuleName string = "hyperdrive-stakewise"

	// Author name as it appears in the descriptor
	AuthorName string = "NodeSet"

	// Fully qualified module name
	FullyQualifiedModuleName string = AuthorName + "/" + ModuleName
)
