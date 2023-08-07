// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package newt implements boil's new command.
package newt

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is the New command configuration.
type Config struct {
	TemplatePath string
	Overwrite    bool
	// EditAfterDefine if true opens the newly defined template
	// after its defined by wizard.
	EditAfterDefine bool
	Config          *boil.Config
}

// Run executes the New command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) {

	var (
		repo boil.Repository
		meta *boil.Metafile
		vars = make(boil.Variables)
	)

	if repo, err = boil.OpenRepository(config.Config.GetRepositoryPath()); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if _, err = repo.OpenMeta(config.TemplatePath); err == nil && !config.Overwrite {
		return fmt.Errorf("template %s already exists", config.TemplatePath)
	}

	meta = boil.NewMetafile(config.Config)
	meta.Name, _, _ = strings.Cut(filepath.Base(config.TemplatePath), "#")
	meta.Path = config.TemplatePath
	if err = boil.NewEditor(config.Config, meta).Wizard(); err != nil {
		return fmt.Errorf("execute wizard: %w", err)
	}
	if err = repo.SaveMeta(meta); err != nil {
		return
	}

	vars.AddNew(boil.Variables{
		"TemplatePath": config.TemplatePath,
	})

	if config.EditAfterDefine {
		return config.Config.ExternalEditor.Execute(vars)
	}

	return nil
}
