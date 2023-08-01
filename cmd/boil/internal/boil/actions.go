// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package boil

import (
	"fmt"
	"os"
	"os/exec"
)

// NewAction returns a new *Action.
func NewAction() *Action {
	return &Action{
		Environment: make(map[string]string),
	}
}

// Action defines some external action to execute via command line.
// See Metafile.Actions for details on Action usage.
type Action struct {
	// Description is the description text of the Action. It's an optional text
	// that should describe the action purpose.
	Description string `json:"description,omitempty"`
	// Program is the path to executable to run.
	Program string `json:"program,omitempty"`
	// Arguments are the arguments to pass to the executable.
	Arguments []string `json:"arguments,omitempty"`
	// WorkDir is the working directory to run the Program from.
	WorkDir string `json:"workDir,omitempty"`
	// Environment is the additional values to set in the Program environment.
	Environment map[string]string `json:"environment,omitempty"`
	// NoFail, if true will not break the execution of the process that ran
	// the Action, but it will generate a warning in the output.
	NoFail bool `json:"noFail,omitempty"`
}

// Execute executes the Action and returns nil on success of error if one occurs.
func (self *Action) Execute(variables Variables) (err error) {

	var args []string
	for _, arg := range self.Arguments {
		args = append(args, variables.ReplacePlaceholders(arg))
	}

	var cmd = exec.Command(
		variables.ReplacePlaceholders(self.Program),
		args...,
	)
	cmd.Dir = variables.ReplacePlaceholders(self.WorkDir)
	for k, v := range self.Environment {
		cmd.Env = append(cmd.Env, k+"="+variables.ReplacePlaceholders(v))
	}
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	if err = cmd.Run(); err != nil && !self.NoFail {
		return fmt.Errorf("action execution failed: %w", err)
	}
	return nil
}

// Actions is a slice of Action with some utilities.
type Actions []*Action

// ExecuteAll executes all actions in self. Returns the error of the first
// action that returns it and stops further execution or nil if no errors occur.
func (self Actions) ExecuteAll(variables Variables) (err error) {
	for _, action := range self {
		if err = action.Execute(variables); err != nil {
			return
		}
	}
	return
}
