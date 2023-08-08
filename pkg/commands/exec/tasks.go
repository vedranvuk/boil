// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package exec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/vedranvuk/boil/pkg/boil"
)

// Tasks is a list of Task.
type Tasks []*Task

// Task defines an execution task to perform for a template.
type Task struct {
	// Metafile is the Template Metafile.
	Metafile *boil.Metafile
	// List is a list of actions to be performed for this template.
	List []*Execute
}

// Execute defines an execution action as part of a exec command task.
type Execute struct {
	// Path is the path of the template file or directory.
	// It is relative to repository root and output directory root.
	Path string
	// Source is path of the template file or dir relative to repo root.
	Source string
	// Target is the absolute path of the target file which will contain Source
	// template output. If the source path had placeholder values they will be
	// replaced with actual values to generate output filenames.
	Target string
	// IsDir wil be true if Source is a directory.
	IsDir bool
}

// PresentPrompts presents a prompt to the user on command line for each of
// the prompts defined in metafiles of all tasks in self, in order as they
// appear in self, depth first. If undeclaredOnly is true only prompts for
// entries not found in variables are presented.
//
// Values are stored in variables under names of Variables they prompt for. If
// undeclaredOnly is true, a variable already defined in variables will not be
// prompted for.
func (self Tasks) PresentPrompts(variables boil.Variables, undeclaredOnly bool) (err error) {

	var (
		ui     = boil.NewInterrogator(os.Stdin, os.Stdout)
		input  string
		exists bool
	)

	ui.Printf("Input variable values.\n")

	for _, template := range self {
		for _, prompt := range template.Metafile.Prompts {
			if _, exists = variables[prompt.Variable]; exists && undeclaredOnly {
				continue
			}
		Repeat:
			if input, err = ui.AskValue(
				fmt.Sprintf("%s %s (%s)",
					template.Metafile.Path,
					prompt.Variable,
					prompt.Description,
				), "", prompt.RegExp,
			); err != nil {
				return err
			}
			if input = strings.TrimSpace(input); !prompt.Optional && input == "" {
				ui.Printf("Variable '%s' may not have an empty value.\n", prompt.Variable)
				goto Repeat
			}
			variables[prompt.Variable] = strings.TrimSpace(input)
		}
	}

	return nil
}

// SetTargetsFromState calculates Execution.Target values of all Tasks
// in self from state, expands any variable placeholders in the process and
// returns nil. If an error occurs it is returned and self may be considered
// invalid in undetermined state.
func (self Tasks) SetTargetsFromState(state *state) (err error) {
	for _, template := range self {
		for _, execution := range template.List {
			execution.Target = filepath.Join(
				state.OutputDir,
				state.Data.ReplaceAll(execution.Path),
			)
		}
	}
	return
}

// CheckForTargetConflicts returns nil if none of the Target paths of all
// defined Tasks in self do not point to an existing file. Otherwise a
// descriptive error is returned.
func (self Tasks) CheckForTargetConflicts() (err error) {
	for _, execGroup := range self {
		for _, exec := range execGroup.List {
			if _, err = os.Stat(exec.Target); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("stat target file: %w", err)
				}
			} else {
				return fmt.Errorf("target file already exists: %s", exec.Target)
			}
		}
	}
	return nil
}

// Validate calls Validate on metafiles of each metafile loaded by each Task in
// self. It returns the first validation error that occurs or nil if all passed.
func (self Tasks) Validate(state *state) (err error) {
	for _, template := range self {
		if err = template.Metafile.Validate(state.Repository); err != nil {
			break
		}
	}
	return
}

// ExecPreParseActions executes all PreParse actions defined in all metafiles
// in the order they are defined, depth first. The first error that occurs from
// any action is returned and execution stopped or nil if everything successed.
func (self Tasks) ExecPreParseActions() (err error) {
	for _, template := range self {
		if err = template.Metafile.ExecPreParseActions(); err != nil {
			return
		}
	}
	return
}

// ExecPreExecuteActions executes all PreExecute actions defined in all
// metafiles in the order they are defined, depth first. The first error that
// occurs from any action is returned and execution stopped or nil if everything
// successed.
func (self Tasks) ExecPreExecuteActions(variables boil.Variables) (err error) {
	for _, template := range self {
		if err = template.Metafile.ExecPreExecuteActions(variables); err != nil {
			return
		}
	}
	return
}

// ExecPostExecuteActions executes all PostExecute actions defined in all
// metafiles in the order they are defined, depth first. The first error that
// occurs from any action is returned and execution stopped or nil if everything
// successed.
func (self Tasks) ExecPostExecuteActions(variables boil.Variables) (err error) {
	for _, template := range self {
		if err = template.Metafile.ExecPostExecuteActions(variables); err != nil {
			return
		}
	}
	return
}

// Execute executes all tasks in self or returns an error.
func (self Tasks) Execute(state *state, print bool) (err error) {

	if state.MakeBackups {
		var id string
		if id, err = boil.CreateBackup(state.OutputDir); err != nil {
			return fmt.Errorf("create target dir backup: %w", err)
		}
		defer func() {
			if err != nil {
				if e := boil.RestoreBackup(id); e != nil {
					err = fmt.Errorf("restore backup failed after error '%w': %w", err, e)
				}
			}
		}()
	}

	for _, exec := range self {
		// Create dirs.
		for _, item := range exec.List {
			if !item.IsDir {
				continue
			}
			if err = os.MkdirAll(item.Target, os.ModePerm); err != nil {
				return fmt.Errorf("error creating target directory %s: %w", item.Target, err)
			}
		}

		// Execute source templates.
		for _, item := range exec.List {
			if item.IsDir {
				continue
			}
			var (
				buf  []byte
				tmpl = template.New(filepath.Base(item.Source)).Funcs(state.Data.Bast.FuncMap())
				file *os.File
			)
			if buf, err = state.Repository.ReadFile(item.Source); err != nil {
				return fmt.Errorf("read template file '%s': %w", item.Source, err)
			}
			if tmpl, err = tmpl.Parse(string(buf)); err != nil {
				return fmt.Errorf("parse template file: %w", err)
			}
			if print {
				fmt.Printf("Template %s\n", tmpl.Name())
				boil.PrintTemplate(tmpl)
			}
			if err = os.MkdirAll(filepath.Dir(item.Target), os.ModePerm); err != nil {
				return fmt.Errorf("create target file dir '%s': %w", filepath.Dir(item.Target), err)
			}
			if file, err = os.Create(item.Target); err != nil {
				return fmt.Errorf("create target file '%s': %w", item.Target, err)
			}
			defer file.Close()
			if err = tmpl.Execute(file, state.Data); err != nil {
				return fmt.Errorf("execute template '%s' into target '%s': %w", item.Source, item.Target, err)
			}
		}
	}
	return nil
}

// Print prints self to stdout.
func (self Tasks) Print() {
	var wr = tabwriter.NewWriter(os.Stdout, 2, 2, 2, 32, 0)
	fmt.Println("Executions:")
	for _, exec := range self {
		fmt.Fprintf(wr, "[Template]\t[Source]\t[Target]\n")
		for _, def := range exec.List {
			fmt.Fprintf(wr, "%s\t%s\t%s\n", exec.Metafile.Path, def.Source, def.Target)
		}
	}
	wr.Flush()
}
