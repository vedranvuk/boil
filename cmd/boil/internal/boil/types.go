package boil

import (
	"strings"
)

// Author defines a template author.
type Author struct {
	// Name is the author name in an arbitrary format.
	Name string
	// Email is the author Email address.
	Email string
	// Homepage is the author's homepage URL.
	Homepage string
}

// Data is the top level structure passed to a template file.
type Data struct {
	// Meta is the metadata of the source template.
	Meta *Metadata
	// Data is the data used by template files.
	Data map[string]string
	// Vars specified on command invocation.
	Vars map[string]string
}

// Vars is a map of variables available to template filemname placeholder
// expansion and from within template files via the .Vars field.
type Vars map[string]string

// ReplaceAll replaces all known placeholders with their actual values.
func (self Vars) ReplaceAll(in string) (out string) {
	out = in
	for k, v := range self {
		out = strings.ReplaceAll(out, "$"+k, v)
	}
	return
}


