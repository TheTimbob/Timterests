package web

import "embed"

// Files embeds the assets directory for serving static files.
//go:embed "assets"
var Files embed.FS
