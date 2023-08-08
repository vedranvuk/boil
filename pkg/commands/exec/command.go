// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package exec implements boil's exec command.
package exec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vedranvuk/boil/pkg/bast"
	"github.com/vedranvuk/boil/pkg/boil"
)

// Config is the Exec command configuration.
type Config struct {
	// TemplatePath is the source template path. During Run() it is adjusted to
	// an absolute path to the Template either inside or outside of repository.
	//
	// If the path is rooted, i.e. starts with "/" the path is treated as an
	// absolute path to a Template and no repository is being loaded or used.
	//
	// If the path is not rooted, the path is treated as a path to a Template
	// relative to the loaded repository.
	//
	// If TemplatePath is an absolute filesystem path it is adjusted to an
	// empty string during Run().
	TemplatePath string

	// OutputDir is the output directory where Template will be executed.
	// If the value is empty the Template will be executed in the current
	// working directory
	//
	// TargetPath is adjusted to an absolute path of OutputDir during Run().
	OutputDir string

	// Overwrite, if true specifies that any file matching a Template output
	// file already existing in the target directory may be overwritten without
	// prompting the user or generating an error.
	Overwrite bool

	// NoExecute if true will not execute any write operations and will
	// instead print out the operations like boil.Config.Verbose was enabled.
	NoExecute bool

	// NoPrompts if true disables prompting the user for variables and will
	// return an error if a variable declared in a prompt was not parsed from
	// the command line.
	NoPrompts bool

	// NoRepository if true does not use a repository when interpreting template
	// paths. Relative paths will be treated as relative to current working
	// directory.
	NoRepository bool

	// EditAfterExec if true opens the output with the editor.
	EditAfterExec bool

	// GoInputs is a list of paths of go files or packages to parse and make
	// their AST available to template files.
	GoInputs []string

	// Vars are variables given by the user on command line.
	// These variables will be available via .Vars template field.
	Vars boil.Variables

	// Config is the loaded program configuration.
	Config *boil.Config
}

// ShouldPrint returns true if Config.Verbose or Config.NoExecute is true.
func (self *Config) ShouldPrint() bool {
	return self.Config.Overrides.Verbose || self.NoExecute
}

// GetRepositoryPath returns the RepositoryPath considering override values.
func (self *Config) GetRepositoryPath() string {
	return self.Config.GetRepositoryPath()
}

// templateData is the top level data structure passed to a template file being 
// executed using text/template engine before being saved to its output.
type templateData struct {
	// Vars contain variables defined via prompts or command line, command
	// specific variables and system variables.
	Vars boil.Variables
	// Bast is the bastard go ast.
	Bast *bast.Bast
}

// newTemplateData returns a new *templateData instance.
func newTemplateData() *templateData {
	return &templateData{
		Vars: make(boil.Variables),
	}
}

// ReplaceAll replaces all known variable placeholders in input string with
// actual values and returns it.
func (self *templateData) ReplaceAll(in string) (out string) {
	return self.Vars.ReplacePlaceholders(in)
}

// state maintains exec command execution.
// It's passed around the files in this package.
type state struct {
	RepositoryPath string
	// TemplatePath is the adjusted path to the template usable by Repository.
	TemplatePath string
	// OutputDir is the adjusted absolute path to the output directory.
	OutputDir string
	// Repository is the loaded Repository.
	Repository boil.Repository
	// Data for Template files, combined from various inputs.
	Data *templateData
	// MakeBackup dictates if backups should be made on execution.
	MakeBackups bool
	// Templates are the Templates to execute.
	Templates Tasks
}

// Run executes the Exec command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) {

	if config.NoExecute {
		fmt.Printf("NoExecute enabled, printing commands instead of executing.\n")
	}

	// Init state to Config values.
	var state = &state{
		RepositoryPath: config.GetRepositoryPath(),
		TemplatePath:   config.TemplatePath,
		OutputDir:      config.OutputDir,
		MakeBackups:    config.Config.ShouldBackup(),
		Data:           newTemplateData(),
	}

	// Determine Input and Output locations.
	if filepath.IsAbs(config.TemplatePath) || config.NoRepository {
		// If TemplatePath is an absolute path open the Template as the
		// Repository and adjust the template path to "current directory"
		// pointing to repository root.
		if path, group, found := strings.Cut(config.TemplatePath, "#"); found {
			state.TemplatePath = ".#" + group
			state.RepositoryPath = path
		} else {
			state.TemplatePath = "."
			state.RepositoryPath = path
		}
		if config.ShouldPrint() {
			fmt.Println("Absolute Template path specified, repository opened at template root.")
		}
	}
	if state.OutputDir, err = filepath.Abs(config.OutputDir); err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	if state.Repository, err = boil.OpenRepository(state.RepositoryPath); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	// Create a Template list, it will contain only the source paths of all
	// referenced template file paths over all referenced templates in a
	// possible group. Outputs are determined later after all variables have
	// been loaded.
	if state.Templates, err = templatesFromState(state); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("not a boil template: %s", config.TemplatePath)
		}
		return fmt.Errorf("enumerate template files for execution: %w", err)
	}

	if err = state.Templates.ExecPreParseActions(); err != nil {
		return fmt.Errorf("pre parse action failed: %w", err)
	}

	// Add vars declared on command line.
	state.Data.Vars.AddNew(config.Vars)
	// Show prompts for variables not satisified early on command line.
	if !config.NoPrompts {
		if err = state.Templates.PresentPrompts(state.Data.Vars, true); err != nil {
			return fmt.Errorf("prompt user for input data: %w", err)
		}
	}
	// Override state variables given on command line with ones given in prompt.
	state.Data.Vars.MaybeSetString(boil.VarOutputDirectory, &state.OutputDir)
	// Append system variables to state vars if not given
	state.Data.Vars.AddNew(boil.Variables{
		boil.VarTemplatePath.String(): state.TemplatePath,
	})

	// Load bast.
	if state.Data.Bast, err = bast.Load(config.GoInputs...); err != nil {
		return fmt.Errorf("process go input files: %w", err)
	}
	if config.ShouldPrint() {
		fmt.Println("Go input:")
		state.Data.Bast.Print(os.Stdout)
	}

	// Now that the vars have been loaded expand variable placeholders in
	// template paths.
	if err = state.Templates.SetTargetsFromState(state); err != nil {
		return fmt.Errorf("expand target file names: %w", err)
	}

	// Validate templates and optionally check for output conflicts.
	if err = state.Templates.Validate(state); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	if !config.Overwrite {
		if err = state.Templates.CheckForTargetConflicts(); err != nil {
			return err
		}
	}

	// Optional verbose output.
	if config.ShouldPrint() {
		fmt.Printf("Repository location: %s\n", state.Repository.Location())
		state.Templates.Print()
		fmt.Println("Templates:")
		for _, m := range state.Templates {
			fmt.Printf("Template %s\n", m.Metafile.Path)
			m.Metafile.Print()
		}
		state.Data.Vars.Print()
	}

	// Exec Pre actions, templates then Post actions. Optionally open output
	// directory in external editor.
	if err = state.Templates.ExecPreExecuteActions(state.Data.Vars); err != nil {
		return fmt.Errorf("pre execute action failed: %w", err)
	}
	if err = state.Templates.Execute(state, config.ShouldPrint()); err != nil {
		return
	}
	if err = state.Templates.ExecPostExecuteActions(state.Data.Vars); err != nil {
		return fmt.Errorf("post execute action failed: %w", err)
	}
	if config.EditAfterExec {
		if err = config.Config.ExternalEditor.Execute(state.Data.Vars); err != nil {
			return
		}
	}

	return nil
}

// templatesFromState returns Templates to be executed from a state. It
// returns empty Templates and an error if the state is invalid, one or more
// template files is missing, any group addresses a missing template or some
// other error.
func templatesFromState(state *state) (templates Tasks, err error) {
	err = produceTemplates(state, state.TemplatePath, &templates)
	return
}

// produceTemplates uses state to recursively construct execs starting
// from path. if the function failes it returns an error.
func produceTemplates(state *state, path string, out *Tasks) (err error) {

	var (
		meta   *boil.Metafile
		group  string
		exists bool
	)

	path, group, _ = strings.Cut(path, "#")

	if meta, err = state.Repository.OpenMeta(path); err != nil {
		return err
	}

	var template = &Task{
		Metafile: meta,
	}

	for _, dir := range meta.Directories {
		template.List = append(template.List, &Execute{
			Path:   dir,
			Source: filepath.Join(path, dir),
			IsDir:  true,
		})
	}

	for _, file := range meta.Files {
		if exists, err = state.Repository.Exists(filepath.Join(path, file)); err != nil {
			return err
		}
		if !exists {
			return fmt.Errorf("template file '%s' does not exist", filepath.Join(path, file))
		}
		template.List = append(template.List, &Execute{
			Path:   file,
			Source: filepath.Join(path, file),
			IsDir:  false,
		})
	}

	*out = append(*out, template)

	if group != "" {
		for _, g := range meta.Groups {
			if g.Name == group {
				continue
			}
			for _, name := range g.Templates {
				if err = produceTemplates(state, filepath.Join(path, name), out); err != nil {
					return
				}
			}
		}
	}

	return nil
}
