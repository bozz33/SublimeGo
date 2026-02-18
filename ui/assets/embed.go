package assets

import "embed"

// FS holds all static assets embedded at compile time.
//
//go:embed css js styles.css
var FS embed.FS
