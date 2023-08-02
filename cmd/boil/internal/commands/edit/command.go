// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package edit implements boil's edit command.
package edit

import (
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is the Edit command configuration.
type Config struct {
	// TemplatePath is the path of the template to edit.
	// It may not contain group names.
	TemplatePath string
	// EditAction specifies the edit sub action.
	EditAction string
	// EditPath is the path to edit by one of fs edit actions.
	EditPath string
	// ForceRemoveNonEmptyDir removal of non-empty directories.
	ForceRemoveNonEmptyDir bool
	// Open the file with editor after touching it.
	EditAfterTouch bool
	// Config is the loaded program configuration.
	Config *boil.Config
}

// Run executes the Edit command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) {

	var (
		repo           boil.Repository
		meta           *boil.Metafile
		vars           boil.Variables
		tmplPath, _, _ = strings.Cut(config.TemplatePath, "#")
		repoPath       = config.Config.GetRepositoryPath()
	)

	if filepath.IsAbs(config.TemplatePath) {
		// If TemplatePath is an absolute path open the Template as the
		// Repository and adjust the template path to "current directory"
		// pointing to repository root.
		repoPath = tmplPath
		tmplPath = "."
		if config.Config.Overrides.Verbose {
			fmt.Println("Absolute Template path specified, repository opened at template root.")
		}
	}

	if repo, err = boil.OpenRepository(repoPath); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if meta, err = repo.OpenMeta(tmplPath); err != nil {
		return fmt.Errorf("template %s not found", config.TemplatePath)
	}

	vars["TemplatePath"] = filepath.Join(repo.Location(), config.TemplatePath)

	switch config.EditAction {
	case "edit":
		return config.Config.ExternalEditor.Execute(vars)
	case "all":
		err = boil.NewEditor(config.Config, meta).EditAll()
	case "info":
		err = boil.NewEditor(config.Config, meta).EditInfo()
	case "files":
		err = boil.NewEditor(config.Config, meta).EditFiles()
	case "dirs":
		err = boil.NewEditor(config.Config, meta).EditDirs()
	case "prompts":
		err = boil.NewEditor(config.Config, meta).EditPrompts()
	case "preparse":
		err = boil.NewEditor(config.Config, meta).EditPreParse()
	case "preexec":
		err = boil.NewEditor(config.Config, meta).EditPreExec()
	case "postexec":
		err = boil.NewEditor(config.Config, meta).EditPostExec()
	case "groups":
		err = boil.NewEditor(config.Config, meta).EditGroups()
	case "file":
		err = openFile(config.EditPath)
	case "touch":
		err = touchFile(config.EditPath)
	case "directory":
		err = openDirectory(config.EditPath)
	case "add":
		err = addDirectory(config.EditPath)
	case "remove":
		err = remove(config.EditPath, config.ForceRemoveNonEmptyDir)
	default:
		panic("unknown edit action")
	}
	if err != nil {
		return
	}
	if config.Config.Overrides.Verbose {
		meta.Print()
	}
	return repo.SaveMeta(meta)

}

var errNotImplemented = errors.New("not implemented")

// openFile opens a file at the path relative to the template directory with
// the editor and returns nil or an error if one occurs.
func openFile(path string) (err error) {
	return errNotImplemented
}

// touchFile creates a new template file at the path relative to the template
// directory and returns nil or an error if one occured.
func touchFile(path string) (err error) {
	return errNotImplemented
}

// deleteFile deletes a template file at the path relative to the template
// directory and returns nil or an error if one occured.
func deleteFile(path string) (err error) {
	return errNotImplemented
}

// openDirectory opens a directory at the path relative to the template
// directory with the editor and returns nil or an error if one occurs.
func openDirectory(path string) (err error) {
	return errNotImplemented
}

// addDirectory creates a new directory at the path relative to the template
// directory and returns nil or an error if one occured. If the directory
// already exists the function is a no-op.
func addDirectory(path string) (err error) {
	return errNotImplemented
}

// removeDirectory deletes a directory at the path relative to the template
// directory and returns nil or an error if one occured. If force is true
// removes even if not empty otherwise generates an error.
func remove(path string, force bool) (err error) {
	return
}
