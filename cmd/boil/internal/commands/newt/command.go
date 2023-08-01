// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package edit implements boil's new command.
package newt

import (
	"fmt"
	"path/filepath"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is the New command configuration.
type Config struct {
	TemplatePath  string
	Overwrite     bool
	Configuration *boil.Config
}

// Run executes the New command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) error { return newState().Run(config) }

// newState returns a new state.
func newState() *state {
	return &state{
		vars: make(boil.Variables),
	}
}

// state is the execution state of the new command.
type state struct {
	config   *Config
	repo     boil.Repository
	metamap  boil.Metamap
	metafile *boil.Metafile
	vars     boil.Variables
}

// Run executes the New command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func (self *state) Run(config *Config) (err error) {
	if self.config = config; self.config == nil {
		return fmt.Errorf("nil config")
	}
	if err = boil.IsValidTemplatePath(config.TemplatePath); err != nil {
		return err
	}
	if self.repo, err = boil.OpenRepository(config.Configuration); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if self.metamap, err = self.repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	if _, err = self.metamap.Metafile(config.TemplatePath); err == nil && !config.Overwrite {
		return fmt.Errorf("template %s already exists", config.TemplatePath)
	}
	if self.metafile, err = self.repo.NewTemplate(config.TemplatePath); err != nil {
		return fmt.Errorf("create new template: %w", err)
	}
	if err = boil.NewEditor(self.config.Configuration, self.metafile).Wizard(); err != nil {
		return fmt.Errorf("execute wizard: %w", err)
	}
	if err = self.repo.SaveTemplate(self.metafile); err != nil {
		return
	}
	self.vars["TemplatePath"] = filepath.Join(self.repo.Location(), config.TemplatePath)
	return self.config.Configuration.ExternalEditor.Execute(self.vars)
}