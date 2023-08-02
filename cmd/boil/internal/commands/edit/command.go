// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package edit implements boil's edit command.
package edit

import (
	"errors"
	"fmt"
	"path/filepath"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is the Edit command configuration.
type Config struct {
	// TemplatePath is the path of the template to edit.
	// It may not contain group names.
	TemplatePath string
	// EditAction specifies the edit sub action.
	EditAction string
	// ForceRemoveNonEmptyDir removal of non-empty directories.
	ForceRemoveNonEmptyDir bool
	// Open the file with editor after touching it.
	EditAfterTouch bool
	// Config is the loaded program configuration.
	Config *boil.Config
}

// Run executes the Edit command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) { return newState().Run(config) }

// newState returns a new state.
func newState() *state {
	return &state{
		vars: make(boil.Variables),
	}
}

// state is the execution state of the edit command.
type state struct {
	repo     boil.Repository
	metafile *boil.Metafile
	vars     boil.Variables
}

// Run executes the Edit command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func (self *state) Run(config *Config) (err error) {

	if self.repo, err = boil.OpenRepository(config.Config); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if self.metafile, err = self.repo.OpenMeta(config.TemplatePath); err != nil {
		return fmt.Errorf("template %s not found", config.TemplatePath)
	}

	self.vars["TemplatePath"] = filepath.Join(self.repo.Location(), config.TemplatePath)

	switch config.EditAction {
	case "edit":
		return config.Config.ExternalEditor.Execute(self.vars)
	case "all":
		err = boil.NewEditor(config.Config, self.metafile).EditAll()
	case "info":
		err = boil.NewEditor(config.Config, self.metafile).EditInfo()
	case "files":
		err = boil.NewEditor(config.Config, self.metafile).EditFiles()
	case "dirs":
		err = boil.NewEditor(config.Config, self.metafile).EditDirs()
	case "prompts":
		err = boil.NewEditor(config.Config, self.metafile).EditPrompts()
	case "preparse":
		err = boil.NewEditor(config.Config, self.metafile).EditPreParse()
	case "preexec":
		err = boil.NewEditor(config.Config, self.metafile).EditPreExec()
	case "postexec":
		err = boil.NewEditor(config.Config, self.metafile).EditPostExec()
	case "groups":
		err = boil.NewEditor(config.Config, self.metafile).EditGroups()
	case "file":
		err = self.openFile(config.TemplatePath)
	case "touch":
		err = self.touchFile(config.TemplatePath)
	case "delete":
		err = self.deleteFile(config.TemplatePath)
	case "directory":
		err = self.openDirectory(config.TemplatePath)
	case "add":
		err = self.addDirectory(config.TemplatePath)
	case "remove":
		err = self.removeDirectory(config.TemplatePath, config.ForceRemoveNonEmptyDir)
	default:
		panic("unknown edit action")
	}
	if err != nil {
		return
	}
	if config.Config.Overrides.Verbose {
		self.metafile.Print()
	}
	return self.repo.SaveMeta(self.metafile)
}

var errNotImplemented = errors.New("not implemented")

// openFile opens a file at the path relative to the template directory with
// the editor and returns nil or an error if one occurs.
func (self *state) openFile(path string) (err error) {
	return errNotImplemented
}

// touchFile creates a new template file at the path relative to the template
// directory and returns nil or an error if one occured.
func (self *state) touchFile(path string) (err error) {
	return errNotImplemented
}

// deleteFile deletes a template file at the path relative to the template
// directory and returns nil or an error if one occured.
func (self *state) deleteFile(path string) (err error) {
	return errNotImplemented
}

// openDirectory opens a directory at the path relative to the template
// directory with the editor and returns nil or an error if one occurs.
func (self *state) openDirectory(path string) (err error) {
	return errNotImplemented
}

// addDirectory creates a new directory at the path relative to the template
// directory and returns nil or an error if one occured. If the directory
// already exists the function is a no-op.
func (self *state) addDirectory(path string) (err error) {
	return errNotImplemented
}

// removeDirectory deletes a directory at the path relative to the template
// directory and returns nil or an error if one occured. If force is true
// removes self even if not empty otherwise generates an error.
func (self *state) removeDirectory(path string, force bool) (err error) {
	return
}
