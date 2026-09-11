package frontend

import "embed"

// Assets embeds the production web UI in the desktop executable.
//go:embed all:dist
var Assets embed.FS
