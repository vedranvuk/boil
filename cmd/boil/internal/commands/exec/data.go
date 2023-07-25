package exec

import (
	"runtime"
	"strings"
)

// Data is the top level data structure passed to a Template file.
type Data struct {
	// Vars is a collection of system variables always present on template
	// execution, generated from environment.
	Vars map[string]string
	// UserVars is a collection of variables given by the user during execution.
	UserVars map[string]string
}

// DataFromConfig returns *Data from config or an error.
func InitConfigData(config *Config) (err error) {
	config.data = &Data{
		Vars:     make(map[string]string),
		UserVars: make(map[string]string),
	}

	// Go Version.
	var (
		va      = strings.Split(strings.TrimPrefix(runtime.Version(), "go"), ".")
		version = va[0] + "." + va[1]
		_       = version
	)
	config.data.Vars["GoVersion"] = version

	return
}

// ReplaceAll replaces all known variable placeholders in input string with
// actual values and returns it.
func (self *Data) ReplaceAll(in string) (out string) {
	// "ProjectName": filepath.Base(config.ModulePath),
	// "TargetDir":   config.TargetDir,
	// "GoVersion":   version,
	// "ModulePath":  config.ModulePath,

	// 	out = in
	// 	for k, v := range self {
	// 		out = strings.ReplaceAll(out, "$"+k, v)
	// 	}
	return in
}
