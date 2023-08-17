// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package newt implements boil's new command.
package newt

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vedranvuk/boil/pkg/boil"
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
		printer  = boil.NewPrinter(os.Stdout)
		repo     boil.Repository
		meta     *boil.Metafile
		data     = boil.NewData()
		tmplPath = config.TemplatePath
		repoPath = config.Config.GetRepositoryPath()
	)

	// Open repository.
	tmplPath, _, _ = strings.Cut(config.TemplatePath, "#")
	if filepath.IsAbs(config.TemplatePath) || config.Config.Overrides.NoRepository {
		// If TemplatePath is an absolute path open the Template as the
		// Repository and adjust the template path to "current directory"
		// pointing to repository root.
		repoPath = tmplPath
		tmplPath = "."
		if config.Config.Overrides.Verbose {
			printer.Printf("Absolute Template path specified, repository opened at template root.\n")
		}
		// Force dirs at repo location for the new template.
		if err = os.MkdirAll(repoPath, os.ModePerm); err != nil {
			return fmt.Errorf("abs template mkdir: %w", err)
		}
	}
	if repo, err = boil.OpenRepository(repoPath); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if _, err = repo.OpenMeta(config.TemplatePath); err == nil && !config.Overwrite {
		return fmt.Errorf("template %s already exists", config.TemplatePath)
	}

	meta = boil.NewMetafile(config.Config)
	meta.Name, _, _ = strings.Cut(filepath.Base(config.TemplatePath), "#")
	meta.Path = tmplPath
	if err = boil.NewEditor(config.Config, meta).Wizard(); err != nil {
		return fmt.Errorf("execute wizard: %w", err)
	}
	if err = repo.SaveMeta(meta); err != nil {
		return
	}

	data.Vars.AddNew(boil.Variables{
		"TemplatePath": config.TemplatePath,
	})

	if config.EditAfterDefine {
		return config.Config.Editor.Execute(data)
	}

	return nil
}
