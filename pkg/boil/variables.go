// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package boil

import (
	"fmt"
	"io"
	"strconv"
	"strings"
)

// Variable defines a known variable.
// A known variable is a variable that oil sets or manipulates but it can be
// set ot overriden by the user in various situations.
// Each one has its magic described.
type Variable int

const (
	// VarTemplatePath is exec command template-path.
	VarTemplatePath Variable = iota

	VarModulePath

	// VarProjectName is used in the context of a project generation by the
	// exec command. It is by default determined from the base of the OutputDir.
	VarProjectName
	// VarWorkingDir is always available.
	VarWorkingDir
	// VarOutputDir is exec command output-dir.
	// It can be set via prompt to override value giveon on command line.
	VarOutputDir

	// VarEditTarget is set by a command.
	VarEditTarget

	// VarAuthor holds the author name.
	VarAuthorName

	VarAuthorEmail

	VarAuthorHomepage
)

// StdVariables is a slice of standard variables.
var StdVariables = []string{
	"TemplatePath",
	"ModulePath",
	"ProjectName",
	"WorkingDir",
	"OutputDir",
	"EditTarget",
	"AuthorName",
	"AuthorEmail",
	"AuthorHomepage",
}

// String reurns string representation of Variable.
func (self Variable) String() string { return StdVariables[self] }

// Variables defines a map of variables keying variable names to their values.
//
// A variable is a value that is available to Template files on execution
// either as data for a Template file being executed with text/template or as
// values when expending placeholders in Template file names.
//
// Variables can be extracted from files, generated by an external
// command or defined by the user on Template execution via command line.
type Variables map[string]any

// ReplacePlaceholders replaces all known variable placeholders in input string
// with actual values and returns it.
//
// A placeholder is a case sensitive variable name prefixed with "$".
func (self Variables) ReplacePlaceholders(in string) (out string) {
	out = in
	for k, v := range self {
		out = strings.ReplaceAll(out, "$"+k, v.(string))
	}
	return out
}

// Exists returns true if variable under name exists.
func (self Variables) Exists(name string) (exists bool) {
	_, exists = self[name]
	return
}

// MaybeSetString sets the value of out to the value of the Variable under key
// if it exists in self and is of type string. Otherwise, out is unmodified.
func (self Variables) MaybeSetString(key Variable, out *string) {
	if key < 0 || int(key) >= len(StdVariables) {
		return
	}
	var val, exists = self[StdVariables[key]]
	if !exists {
		return
	}
	var str, ok = val.(string)
	if !ok {
		return
	}
	*out = str
}

// AddNew adds variables that do not exist in self to self and returns self.
func (self Variables) AddNew(variables Variables) Variables {
	for k, v := range variables {
		if _, exists := self[k]; exists {
			continue
		}
		self[k] = v
	}
	return self
}

// SetAssignments adds assignments to self overwriting any existing entries
// where an assignment is a string in a "key=value" format.
//
// Value is unquoted if quoted. If an entry is found that is not in proper
// format an error is returned that describes the offending entry and no
// variables are added to self or in case of any other error.
func (self Variables) SetAssignments(assignments ...string) (err error) {
	var (
		temp     = make(Variables)
		key, val string
		valid    bool
	)
	for _, assignment := range assignments {
		if key, val, valid = strings.Cut(assignment, "="); !valid {
			return fmt.Errorf("invalid variable format %s", assignments)
		}
		if val == "" {
			continue
		}
		if strings.HasPrefix(val, "\"") && strings.HasSuffix(val, "\"") {
			if val, err = strconv.Unquote(val); err != nil {
				return fmt.Errorf("unquote var value: %w", err)
			}
		}
		temp[key] = val
	}
	for k, v := range temp {
		self[k] = v
	}
	return
}

func (self Variables) Print(wr io.Writer) {
	if len(self) == 0 {
		return
	}
	fmt.Println("Variables:")
	for k, v := range self {
		fmt.Fprintf(wr, "%s\t%v\n", k, v)
	}
}
