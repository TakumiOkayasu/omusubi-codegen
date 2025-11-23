package generator

import (
	"embed"
)

//go:embed templates/*.tmpl
var templatesFS embed.FS
