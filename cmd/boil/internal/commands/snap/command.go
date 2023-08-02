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

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is the SNap command configuration.
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

	// Configuration is the loaded program configuration.
	Configuration *boil.Config
}

// Run executes the SNapshot command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) { return newState().Run(config) }

func newState() *state { return &state{} }

type state struct {
	config   *Config
	repo     boil.Repository
	metamap  boil.Metamap
	metafile *boil.Metafile
	source   string
}

// Run executes the Snap command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func (self *state) Run(config *Config) (err error) {

	// Checks
	if self.config = config; self.config == nil {
		return fmt.Errorf("nil config")
	}

	// Open repository and get its metamap, check template exists.
	if self.repo, err = boil.OpenRepository(config.Configuration); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if self.metamap, err = self.repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	if _, err = self.metamap.Metafile(config.TemplatePath); err == nil && !config.Overwrite {
		return fmt.Errorf("template %s already exists", config.TemplatePath)
	}

	// Create new metadata, set base author info and file and dir list.
	self.metafile = boil.NewMetafile(config.Configuration)
	if self.source, err = filepath.Abs(config.SourcePath); err != nil {
		return fmt.Errorf("get absolute source path: %w", err)
	}
	var fi fs.FileInfo
	if fi, err = os.Stat(self.source); err != nil {
		return fmt.Errorf("stat source: %w", err)
	} else if fi.IsDir() {
		if err = filepath.WalkDir(self.source, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if path = strings.TrimPrefix(strings.TrimPrefix(path, self.source), "/"); path == "" {
				return nil
			}
			if strings.ToLower(path) == boil.MetafileName {
				return nil
			}
			if d.IsDir() {
				self.metafile.Directories = append(self.metafile.Directories, path)
			} else {
				self.metafile.Files = append(self.metafile.Files, path)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("enumerate source directory: %w", err)
		}
	} else {
		self.metafile.Files = append(self.metafile.Files, self.source)
	}
	self.metafile.Author.Name = self.config.Configuration.DefaultAuthor.Name
	self.metafile.Author.Email = self.config.Configuration.DefaultAuthor.Email
	self.metafile.Author.Homepage = self.config.Configuration.DefaultAuthor.Homepage

	// Template wizard
	if config.Wizard {
		if err = boil.NewEditor(self.config.Configuration, self.metafile).Wizard(); err != nil {
			return fmt.Errorf("execute wizard: %w", err)
		}
	}

	// Check existing template files
	if !config.Overwrite {
		var exists bool
		for _, file := range self.metafile.Files {
			if exists, err = self.repo.Exists(file); err != nil {
				return err
			}
			if exists {
				return fmt.Errorf("template file '%s' already exists", file)
			}
		}
	}

	// Verbose
	if config.Configuration.Overrides.Verbose {
		fmt.Printf("Abs source path:     %s\n", self.source)
		fmt.Printf("Template path:       %s\n", self.config.SourcePath)
		fmt.Printf("Overwrite Template:  %t\n", self.config.Overwrite)
		fmt.Printf("Repository location: %s\n", self.repo.Location())
		fmt.Println()
		fmt.Printf("Metafile:")
		self.metafile.Print()
	}

	if err = self.repo.SaveMeta(self.metafile); err != nil {
		return
	}

	// Create template directories
	for _, dir := range self.metafile.Directories {
		dir = filepath.Join(self.config.TemplatePath, dir)

		if config.Configuration.Overrides.Verbose {
			fmt.Printf("Create template directory: '%s'\n", dir)
		}

		if err = self.repo.Mkdir(dir); err != nil {
			return fmt.Errorf("create template dir: %w", err)
		}
	}

	// Create and copy template files
	for _, file := range self.metafile.Files {

		var (
			data  []byte
			inFn  = filepath.Join(self.source, file)
			outFn = filepath.Join(self.config.TemplatePath, file)
		)

		if config.Configuration.Overrides.Verbose {
			fmt.Printf("Copy %s to %s\n", inFn, outFn)
		}

		if data, err = os.ReadFile(inFn); err != nil {
			return fmt.Errorf("read input file %w", err)
		}
		if err = self.repo.WriteFile(outFn, data); err != nil {
			return fmt.Errorf("write template file: %w", err)
		}
	}

	return
}
