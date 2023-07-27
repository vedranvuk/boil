package boil

import (
	"os"
	"os/exec"
)

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
	NoFail bool
}

// Actions is a slice of Action with some utilities.
type Actions []*Action

// ExecuteAll executes all actions in self. Returns the error of the first
// action that returns it and stops further execution or nil if no errors occur.
func (self Actions) ExecuteAll(variables Variables) (err error) {

	var (
		cmd  *exec.Cmd
		args []string
	)

	for _, action := range self {

		for _, arg := range action.Arguments {
			args = append(args, variables.ReplaceAll(arg))
		}

		cmd = exec.Command(
			variables.ReplaceAll(action.Program),
			args...,
		)
		cmd.Dir = variables.ReplaceAll(action.WorkDir)
		cmd.Stdout = os.Stdout
		cmd.Stderr = os.Stderr

		for k, v := range action.Environment {
			cmd.Env = append(cmd.Env, k+"="+variables.ReplaceAll(v))
		}

		err = cmd.Run()
	}
	return
}
