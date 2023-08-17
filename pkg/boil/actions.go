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

// Execute executes the Action and returns nil on success or an error.
// It expands any template tokens in self definition using data.
func (self *Action) Execute(data *Data) (err error) {

	var (
		prog string
		args []string
	)
	if prog, err = ExecuteTemplateString(self.Program, data); err != nil {
		return fmt.Errorf("expand program: %w", err)

	}
	for _, arg := range self.Arguments {
		if arg, err = ExecuteTemplateString(arg, data); err != nil {
			return fmt.Errorf("expand argument %s: %w", arg, err)
		}
		args = append(args, arg)
	}

	var cmd = exec.Command(
		prog,
		args...,
	)
	if cmd.Dir, err = ExecuteTemplateString(self.WorkDir, data); err != nil {
		return fmt.Errorf("expand workdir: %w", err)
	}
	for k, v := range self.Environment {
		if v, err = ExecuteTemplateString(v, data); err != nil {
			return fmt.Errorf("expand env: %w", err)
		}
		cmd.Env = append(cmd.Env, k+"="+v)
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
func (self Actions) ExecuteAll(data *Data) (err error) {
	for _, action := range self {
		if err = action.Execute(data); err != nil {
			return
		}
	}
	return
}
