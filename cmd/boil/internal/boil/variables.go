package boil

import "strings"

// Variables defines a map of variables keying variable names to their values.
//
// A variable is a value that is available to Template files on execution
// either as data for a Template file being executed with text/template or as
// values when expending placeholders in Template file names.
//
// Variables can be extracted from files, generated by an external
// command or defined by the user on Template execution via command line.
type Variables map[string]any

// ReplaceAll replaces all known variable placeholders in input string with
// actual values and returns it.
func (self Variables) ReplaceAll(in string) (out string) {
	out = in
	for k, v := range self {
		out = strings.ReplaceAll(out, "$"+k, v.(string))
	}
	return out
}
