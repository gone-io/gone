package templates

import "embed"

//go:embed *
//go:embed */.dockerignore
//go:embed */.gitignore
var F embed.FS
