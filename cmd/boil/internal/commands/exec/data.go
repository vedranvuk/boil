package exec

import (
	"fmt"
	"runtime"
	"strings"
)

// VarMap is a map of variables.
type VarMap map[string]any

// Data is the top level data structure passed to a Template file.
type Data struct {
	// Vars is a collection of system variables always present on template
	// execution, generated from environment.
	Vars VarMap
}

// NeData returns new *Data initialized with standard variables or an error if
// one occurs. If merge is not nil it is added to the resulting Data.
func NewData() *Data {
	return &Data{
		Vars: make(VarMap),
	}
}

// AddVar adds a variable and returns nil on success or an error if a variable
// under specified key already exists.
func (self *Data) AddVar(key string, value any) error {
	if _, exists := self.Vars[key]; exists {
		return fmt.Errorf("variable %s already registered", key)
	}
	self.Vars[key] = value
	return nil
}

// InitStandardVars initializes values of standard variables.
func (self *Data) InitStandardVars(state *state) error {
	// Go Version.
	var (
		va      = strings.Split(strings.TrimPrefix(runtime.Version(), "go"), ".")
		version = va[0] + "." + va[1]
	)
	self.Vars["GoVersion"] = version

	return nil
}

// MergeVars merges vars to self.Vars or returns an error.
func (self *Data) MergeVars(vars VarMap) error {
	var exists bool
	for k, v := range vars {
		if _, exists = self.Vars[k]; exists {
			return fmt.Errorf("duplicate variable '%s'", k)
		}
		self.Vars[k] = v
	}
	return nil
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
