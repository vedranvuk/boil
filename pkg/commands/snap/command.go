// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package snap implements boil's snap command.
package snap

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/vedranvuk/boil/pkg/boil"
)

// Config is the Snap command configuration.
type Config struct {
	// TemplatePath is the path under which the Template will be stored
	// relative to the loaded repository root.
	TemplatePath string

	// SourcePath is an optional path to the source directory or file.
	// If ommitted a snapshot of the current directory is created.
	SourcePath string

	// Wizard specifies if a template wizard should be used.
	Wizard bool

	// Force overwriting template if it already exists.
	Overwrite bool

	// Config is the loaded program configuration.
	Config *boil.Config
}

// Run executes the Snap command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) {

	var (
		repo     boil.Repository
		meta     *boil.Metafile
		source   string
		printer  = boil.NewPrinter(os.Stdout)
		tmplPath = config.TemplatePath
		repoPath = config.Config.GetRepositoryPath()
		isAbs    bool
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
		isAbs = true
	}
	if repo, err = boil.OpenRepository(repoPath); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if _, err = repo.OpenMeta(tmplPath); err == nil && !config.Overwrite {
		return fmt.Errorf("template %s already exists", config.TemplatePath)
	}

	// Init new metafile.
	meta = boil.NewMetafile(config.Config)
	meta.Name, _, _ = strings.Cut(filepath.Base(config.TemplatePath), "#")
	meta.Path = tmplPath

	// Determine abs source.
	if source, err = filepath.Abs(config.SourcePath); err != nil {
		return fmt.Errorf("get absolute source path: %w", err)
	}

	// Enum source files and dirs into meta.
	var fi fs.FileInfo
	if fi, err = os.Stat(source); err != nil {
		return fmt.Errorf("stat source: %w", err)
	} else if fi.IsDir() {
		if err = filepath.WalkDir(source, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path, err = filepath.Rel(source, path); err != nil {
				return err
			}
			if path == "." || path == strings.ToLower(boil.MetafileName) {
				return nil
			}
			if d.IsDir() {
				meta.Directories = append(meta.Directories, path)
			} else {
				meta.Files = append(meta.Files, path)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("enumerate source directory: %w", err)
		}
	} else {
		meta.Files = append(meta.Files, source)
	}

	// Optional template wizard then save.
	if config.Wizard {
		if err = boil.NewEditor(config.Config, meta).Wizard(); err != nil {
			return fmt.Errorf("execute wizard: %w", err)
		}
	}
	if err = repo.SaveMeta(meta); err != nil {
		return
	}

	// Check existing template files
	if !config.Overwrite {
		var exists bool
		for _, file := range meta.Files {
			if exists, err = repo.Exists(file); err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("template file '%s' already exists", file)
			}
		}
	}

	// Verbose
	if config.Config.Overrides.Verbose {
		printer.Printf("Abs source path:     %s\n", source)
		printer.Printf("Template path:       %s\n", config.SourcePath)
		printer.Printf("Overwrite Template:  %t\n", config.Overwrite)
		printer.Printf("Repository location: %s\n", repo.Location())
		printer.Printf("\n")
		printer.Printf("Metafile:")
		meta.Print(printer)
	}

	// Create template directories
	for _, dir := range meta.Directories {
		dir = filepath.Join(tmplPath, dir)
		if config.Config.Overrides.Verbose {
			printer.Printf("Create template directory: '%s'\n", dir)
		}
		if err = repo.Mkdir(dir); err != nil {
			return fmt.Errorf("create template dir: %w", err)
		}
	}

	// Create and copy template files
	for _, file := range meta.Files {
		var (
			data  []byte
			inFn  = filepath.Join(source, file)
			outFn = filepath.Join(tmplPath, file)
		)
		if config.Config.Overrides.Verbose {
			printer.Printf("Copy %s to %s\n", inFn, outFn)
		}
		if data, err = os.ReadFile(inFn); err != nil {
			return fmt.Errorf("read input file %w", err)
		}
		if isAbs {
			if err = repo.Mkdir(filepath.Dir(outFn)); err != nil {
				return fmt.Errorf("create template file dir: %w", err)
			}
		}
		if err = repo.WriteFile(outFn, data); err != nil {
			return fmt.Errorf("write template file: %w", err)
		}
	}

	return
}
