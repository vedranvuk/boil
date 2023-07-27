package exec

import (
	"fmt"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Data is the top level data structure passed to a Template file.
type Data struct {
	// Vars is a collection of system variables always present on template
	// execution, generated from environment.
	Vars boil.Variables
}

// NeData returns new *Data initialized with standard variables or an error if
// one occurs. If merge is not nil it is added to the resulting Data.
func NewData() *Data {
	return &Data{
		Vars: make(boil.Variables),
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
	self.Vars["OutputDirectory"] = state.TargetDir
	self.Vars["TemplatePath"] = state.TemplatePath
	return nil
}

// MergeVars merges vars to self.Vars or returns an error.
func (self *Data) MergeVars(vars boil.Variables) error {
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
	return self.Vars.ReplaceAll(in)
}
