// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package info implements boil's info command.
package info

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/vedranvuk/boil/pkg/boil"
)

// Config is the Info command configuration.
type Config struct {
	TemplatePath string
	// Config is the loaded program configuration.
	Config *boil.Config
}

// Run executes the Info command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) {

	var (
		repo     boil.Repository
		meta     *boil.Metafile
		printer  = boil.NewPrinter(os.Stdout)
		repoPath = config.Config.GetRepositoryPath()
		tmplPath string
	)

	tmplPath, _, _ = strings.Cut(config.TemplatePath, "#")
	if filepath.IsAbs(config.TemplatePath) || config.Config.Overrides.NoRepository {
		// If TemplatePath is an absolute path open the Template as the
		// Repository and adjust the template path to "current directory"
		// pointing to repository root.
		repoPath = tmplPath
		tmplPath = "."
		if config.Config.Overrides.Verbose {
			printer.Printf("Absolute Template path specified, repository opened at template root.")
		}
	}

	// Open repository and get its metamap, check template exists.
	if repo, err = boil.OpenRepository(repoPath); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if meta, err = repo.OpenMeta(tmplPath); err != nil {
		return fmt.Errorf("template %s not found", config.TemplatePath)
	}

	meta.Print(printer)
	printer.Flush()

	return nil
}
