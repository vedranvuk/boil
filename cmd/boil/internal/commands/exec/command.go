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

	// Vars are variables given by the user on command line.
	// These variables will be available via .Vars template field.
	Vars boil.Variables

	// Configuration is the loaded program configuration.
	Configuration *boil.Config
}

// ShouldPrint returns true if Config.Verbose or Config.NoExecute is true.
func (self *Config) ShouldPrint() bool {
	return self.Configuration.Overrides.Verbose || self.NoExecute
}

// GetRepositoryPath returns the RepositoryPath considering override values.
func (self *Config) GetRepositoryPath() string {
	return self.Configuration.GetRepositoryPath()
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
	// Metamap is the metamap of the loaded repository.
	Metamap boil.Metamap
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

	var state = &state{
		RepositoryPath: config.GetRepositoryPath(),
		TemplatePath:   config.TemplatePath,
		OutputDir:      config.OutputDir,
		MakeBackups:    config.Configuration.ShouldBackup(),
		Data:           NewData(),
	}

	if config.NoExecute {
		fmt.Printf("NoExecute enabled, printing commands instead of executing.\n")
	}

	if filepath.IsAbs(config.TemplatePath) {
		// If TemplatePath is an absolute path open the Template as the
		// Repository and adjust the template path to "current directory"
		// pointing to repository root.
		state.RepositoryPath = config.TemplatePath
		state.TemplatePath = "."
		if config.ShouldPrint() {
			fmt.Println("Absolute Template path specified, repository opened at template root.")
		}
	}

	if state.Repository, err = boil.OpenRepository(config.Configuration); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	if state.Metamap, err = state.Repository.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}

	if state.OutputDir, err = filepath.Abs(config.OutputDir); err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	if err = state.Data.InitStandardVars(state); err != nil {
		return fmt.Errorf("initialize data: %w", err)
	}

	if err = state.Data.MergeVars(config.Vars); err != nil {
		return fmt.Errorf("load user variables: %w", err)
	}

	if state.Templates, err = GetTemplatesForState(state); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("not a boil template: %s", config.TemplatePath)
		}
		return fmt.Errorf("enumerate template files for execution: %w", err)
	}

	if err = state.Templates.ExecPreParseActions(); err != nil {
		return fmt.Errorf("pre parse action failed: %w", err)
	}

	if !config.NoPrompts {
		if err = state.Templates.PresentPrompts(state.Data.Vars, true); err != nil {
			return fmt.Errorf("prompt user for input data: %w", err)
		}
	} else {
		if err = state.Templates.ValidateVariablesFromPrompts(state.Data.Vars); err != nil {
			return err
		}
	}

	if err = state.Templates.ExpandExecutionTargets(state.Data); err != nil {
		return fmt.Errorf("expand target file names: %w", err)
	}

	if err = state.Templates.Validate(state); err != nil {
		return fmt.Errorf("validation failed: %w", err)
	}

	if config.ShouldPrint() {
		fmt.Printf("Repository location: %s\n", state.Repository.Location())
		state.Metamap.Print()
		state.Templates.Print()
	}

	if !config.Overwrite {
		if err = state.Templates.CheckForTargetConflicts(); err != nil {
			return err
		}
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

	return nil
}
