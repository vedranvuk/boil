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

	"github.com/vedranvuk/boil/cmd/boil/internal/bast"
	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
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

// state is the Exec command state.
type state struct {
	RepositoryPath string
	// TemplatePath is the adjusted path to the template usable by Repository.
	TemplatePath string
	// OutputDir is the adjusted absolute path to the output directory.
	OutputDir string
	// Repository is the loaded Repository.
	Repository boil.Repository
	// Data for Template files, combined from various inputs.
	Data *Data
	// MakeBackup dictates if backups should be made on execution.
	MakeBackups bool
	// Templates are the Templates to execute.
	Templates Templates
}

// Run executes the Exec command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) {

	if config.NoExecute {
		fmt.Printf("NoExecute enabled, printing commands instead of executing.\n")
	}

	var state = &state{
		RepositoryPath: config.GetRepositoryPath(),
		TemplatePath:   config.TemplatePath,
		OutputDir:      config.OutputDir,
		MakeBackups:    config.Config.ShouldBackup(),
		Data:           NewData(),
	}

	// Parse template path and putput dir
	if filepath.IsAbs(config.TemplatePath) {
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

	// Open repo, get a list of templates to execute, run pre-parse actions.
	if state.Repository, err = boil.OpenRepository(state.RepositoryPath); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if state.Templates, err = GetSourceTemplates(state); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("not a boil template: %s", config.TemplatePath)
		}
		return fmt.Errorf("enumerate template files for execution: %w", err)
	}
	if err = state.Templates.ExecPreParseActions(); err != nil {
		return fmt.Errorf("pre parse action failed: %w", err)
	}

	// Add vars declared on command line to state vars.
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
		return fmt.Errorf("error processing go input files: %w", err)
	}

	// Expand variable placeholders in paths.
	if err = state.Templates.DetermineTemplateTargets(state); err != nil {
		return fmt.Errorf("expand target file names: %w", err)
	}
	// Validate templates, check for conflicts.
	if err = state.Templates.Validate(state); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}
	if !config.Overwrite {
		if err = state.Templates.CheckForTargetConflicts(); err != nil {
			return err
		}
	}
	// Print, run pre actions, exec template files, post actions and edit.
	if config.ShouldPrint() {
		fmt.Printf("Repository location: %s\n", state.Repository.Location())
		state.Templates.Print()
	}
	if err = state.Templates.ExecPreExecuteActions(state.Data.Vars); err != nil {
		return fmt.Errorf("pre execute action failed: %w", err)
	}
	if err = state.Templates.Execute(state); err != nil {
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
