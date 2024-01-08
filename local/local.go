package local

import "embed"

//go:embed clients/*.yaml
var Clients embed.FS

//go:embed vaults/*.env
var Vaults embed.FS

//go:embed compose.yaml
var Compose []byte

//go:embed compose.internal.yaml
var ComposeInternal []byte
