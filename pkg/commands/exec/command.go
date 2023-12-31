// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package exec implements boil's exec command.
package exec

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/vedranvuk/bast/pkg/bast"
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

	// NoMetadata if true disables parsing template metadata and copies the
	// source template files recursively to output directory. This disables
	// groups and prompts but the variable system still works via command line.
	NoMetadata bool

	// EditAfterExec if true opens the output with the editor.
	EditAfterExec bool

	// GoInputs is a list of paths of go files or packages to parse and make
	// their AST available to template files.
	GoInputs []string

	// JsonInputs is a list of paths of json files to parse and make available
	// to template files.
	JsonInputs []string

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
	Data *boil.Data
	// MakeBackup dictates if backups should be made on execution.
	MakeBackups bool
	// Tasks are the Tasks to execute.
	Tasks Tasks
}

// Run executes the Exec command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) {

	var printer = boil.NewPrinter(os.Stdout)

	if config.NoExecute {
		printer.Printf("NoExecute enabled, printing commands instead of executing.\n")
	}

	// Init state to Config values.
	var state = &state{
		RepositoryPath: config.GetRepositoryPath(),
		TemplatePath:   config.TemplatePath,
		OutputDir:      config.OutputDir,
		MakeBackups:    config.Config.ShouldBackup(),
		Data:           boil.NewData(),
	}

	// Determine repository and template paths then open repository.
	if !boil.IsRepoPath(config.TemplatePath) || config.Config.Overrides.NoRepository {
		// If TemplatePath is an absolute path or no repository use is forced
		// open the Template directory as Repository and adjust the template
		// path to "current directory" pointing to repository root.
		if path, group, found := strings.Cut(config.TemplatePath, "#"); found {
			state.TemplatePath = ".#" + group
			state.RepositoryPath = path
		} else {
			state.TemplatePath = "."
			state.RepositoryPath = path
		}
		if config.ShouldPrint() && config.Config.Overrides.NoRepository {
			printer.Printf("No repository mode.\n")
		}
	}
	if state.Repository, err = boil.OpenRepository(state.RepositoryPath); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	// Determine absolute output path.
	if state.OutputDir, err = filepath.Abs(config.OutputDir); err != nil {
		return fmt.Errorf("get absolute output path: %w", err)
	}
	// Produce execution tasks depending on execution mode.
	switch config.NoMetadata {
	case false:
		// Create a Template list, it will contain only the source paths of all
		// referenced template file paths over all referenced templates in a
		// possible group. Outputs are determined later after all variables have
		// been loaded.
		if state.Tasks, err = tasksFromMetafile(state.Repository, state.TemplatePath); err != nil {
			if errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("not a boil template: %s", config.TemplatePath)
			}
			return fmt.Errorf("enumerate template files for execution: %w", err)
		}
	case true:
		if state.Tasks, err = tasksFromWalk(state.Repository, state.TemplatePath); err != nil {
			return fmt.Errorf("enumerate template files for execution: %w", err)
		}
	}
	// Exec pre parse actions.
	if err = state.Tasks.ExecPreParseActions(); err != nil {
		return fmt.Errorf("pre parse action failed: %w", err)
	}
	// Load Data.
	if state.Data, err = boil.DataFromInputs(config.Vars, config.GoInputs, config.JsonInputs); err != nil {
		return fmt.Errorf("load data: %w", err)
	}

	// TODO: Clean up variable expansion.

	// Set vars declared on command line as initial state vars before showing
	// prompts so variables declared in prompts that have had their value set
	// early via command line will not be prompted for.
	state.Data.Vars[boil.VarOutputDir.String()] = state.OutputDir
	state.Data.Vars[boil.VarTemplatePath.String()] = state.TemplatePath
	state.Data.Vars[boil.VarAuthorName.String()] = config.Config.Author.Name
	state.Data.Vars[boil.VarAuthorEmail.String()] = config.Config.Author.Email
	state.Data.Vars[boil.VarAuthorHomepage.String()] = config.Config.Author.Homepage
	if !config.NoPrompts && !config.NoMetadata {
		if err = state.Tasks.PresentPrompts(
			state,
			func(p *boil.Prompt) (def string, present bool) {
				present = true
				switch p.Variable {
				case boil.VarProjectName.String():
					// Set default value to base of output dir in vars.
					// if current dir, set from state.
					// if again current dir, use a magic name.
					if def = filepath.Base(state.Data.StringVar(boil.VarOutputDir.String())); def != "" && def != "." {
						return
					}
					if def = filepath.Base(state.OutputDir); def == "." {
						def = "new_project"
					}
				case boil.VarModulePath.String():
					// Set default value to config.ModulePrefix + ProjectName.
					if def = state.Data.StringVar(boil.VarProjectName.String()); def == "" || def == "." {
						if def = filepath.Base(state.OutputDir); def == "." {
							def = "new_project"
						}
					}
					def = filepath.Join(config.Config.Author.ModulePrefix, def)
				case boil.VarOutputDir.String():
					def = state.OutputDir
				}
				return
			},
		); err != nil {
			return fmt.Errorf("prompt user: %w", err)
		}
	}
	// If OutputDir was set via Prompt update state.OutputDir.
	if state.Data.Vars.Exists(boil.VarOutputDir.String()) {
		if state.Data.Vars[boil.VarOutputDir.String()], err = filepath.Abs(state.Data.StringVar(boil.VarOutputDir.String())); err != nil {
			return fmt.Errorf("state output dir: %w", err)
		}
		state.OutputDir = state.Data.StringVar(boil.VarOutputDir.String())
	}
	// Optionally print Bast.
	if config.ShouldPrint() && len(state.Data.Bast.Packages) > 0 {
		printer.Printf("Go input:\n")
		bast.Print(os.Stdout, state.Data.Bast)
	}
	// Now that the vars have been loaded expand variable placeholders in
	// template paths.
	if err = state.Tasks.SetTargetsFromState(state); err != nil {
		return fmt.Errorf("expand target file names: %w", err)
	}
	// Validate templates and optionally check for output conflicts.
	if !config.NoMetadata {
		if err = state.Tasks.Validate(state); err != nil {
			return fmt.Errorf("validation failed: %w", err)
		}
	}
	if !config.Overwrite {
		if err = state.Tasks.CheckForTargetConflicts(); err != nil {
			return err
		}
	}
	// Optional verbose output.
	if config.ShouldPrint() {
		printer.Printf("Repository location: %s\n", state.Repository.Location())
		state.Tasks.Print(printer)
		state.Data.Vars.Print(printer)
	}
	// Exec Pre actions, templates then Post actions. Optionally open output
	// directory in external editor.
	if err = state.Tasks.ExecPreExecuteActions(state.Data); err != nil {
		return fmt.Errorf("pre execute action failed: %w", err)
	}
	if err = state.Tasks.Execute(state, config.ShouldPrint()); err != nil {
		return
	}
	if err = state.Tasks.ExecPostExecuteActions(state.Data); err != nil {
		return fmt.Errorf("post execute action failed: %w", err)
	}
	if config.EditAfterExec {
		state.Data.Vars.AddNew(boil.Variables{
			boil.VarEditTarget.String(): state.OutputDir,
		})
		if err = config.Config.Editor.Execute(state.Data); err != nil {
			return
		}
	}
	return nil
}

// tasksFromMetafile returns Templates to be executed from a state. It
// returns empty Templates and an error if the state is invalid, one or more
// template files is missing, any group addresses a missing template or some
// other error.
func tasksFromMetafile(repo boil.Repository, path string) (templates Tasks, err error) {
	err = produceTasksFromMetafile(repo, path, &templates)
	return
}

// produceTasksFromMetafile uses state to recursively construct execs starting
// from path. if the function failes it returns an error.
func produceTasksFromMetafile(repo boil.Repository, path string, out *Tasks) (err error) {

	var (
		meta   *boil.Metafile
		group  string
		exists bool
	)

	path, group, _ = strings.Cut(path, "#")

	if meta, err = repo.OpenMeta(path); err != nil {
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
		if exists, err = repo.Exists(filepath.Join(path, file)); err != nil {
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
				if err = produceTasksFromMetafile(repo, filepath.Join(path, name), out); err != nil {
					return
				}
			}
		}
	}

	return nil
}

// tasksFromWalk returns Templates to be executed from walking the repo starting
// at the root directory or an error if one occured.
// It returns a single tasks that holds all Execs.
func tasksFromWalk(repo boil.Repository, root string) (out Tasks, err error) {
	var task = new(Task)
	if err = repo.WalkDir(root, func(path string, d fs.DirEntry, err error) error {
		if root == path {
			return nil
		}
		var (
			exe = new(Execute)
			rel string
		)
		if rel, err = filepath.Rel(root, path); err != nil {
			return err
		}
		exe.Path = rel
		exe.Source = path
		exe.IsDir = d.IsDir()
		task.List = append(task.List, exe)
		return nil
	}); err != nil {
		return nil, fmt.Errorf("enumerate source files: %w", err)
	}
	out = append(out, task)
	return
}
