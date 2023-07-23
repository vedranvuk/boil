package exec

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"
)

// Execution defines an execution of a single template file.
type Execution struct {
	// Source is the absolute path of the source file which is to be executed
	// onto Target path.
	// The path may be contain placeholder values.
	Source string
	// Target is the absolute path of the target file which will contain Source
	// template output. If the source path had placeholder values they will be
	// replaced with actual values.
	Target string
}

// Data is the top level data structure passed to a Template file.
type Data struct {
	// Vars is a collection of system variables always present on template
	// execution, generated from environment.
	Vars map[string]string
	// UserVars is a collection of variables given by the user during execution.
	UserVars map[string]string
}

// Execute executes a FileCopy operation or returns an error.
func (self *Execution) Execute(data interface{}) error {

	var (
		err  error
		buf  []byte
		tmpl *template.Template
		file *os.File
	)

	if buf, err = ioutil.ReadFile(self.Source); err != nil {
		return fmt.Errorf("read source template '%s': %w", self.Source, err)
	}

	if tmpl, err = template.New(filepath.Base(self.Source)).Parse(string(buf)); err != nil {
		return fmt.Errorf("parse source template '%s': %w", self.Source, err)
	}

	if file, err = os.Create(self.Target); err != nil {
		return fmt.Errorf("create target file '%s': %w", self.Target, err)
	}
	defer file.Close()

	if err = tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("execute template '%s' to '%s': %w", self.Source, self.Target, err)
	}

	return nil
}

// Executions is a list of executions needed to execute a Boil template.
type Executions []*Execution

// PrepareExecutions prepares a list of executions from an Exec Config
// or returns an error if one occurs.
func PrepareExecutions(config *Config) (list []*Execution, err error) {
	return nil, nil
}
