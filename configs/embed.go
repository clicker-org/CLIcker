// Package configs provides embedded world configuration data for use by world
// implementations. Keeping the embed here avoids the go:embed restriction on
// paths containing "..".
package configs

import _ "embed"

// TerraToml is the embedded configs/worlds/terra.toml configuration.
//go:embed worlds/terra.toml
var TerraToml []byte
