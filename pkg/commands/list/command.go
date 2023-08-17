// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package list implements boil's list command.
package list

import (
	"fmt"
	"os"
	"strings"

	"github.com/vedranvuk/boil/pkg/boil"
)

// Config is the List command configuration.
type Config struct {
	// Prefix is the path prefix at which to start listing.
	Prefix string
	// Config is the loaded program configuration.
	Config *boil.Config
}

// Run executes the List command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) {

	var (
		repo    boil.Repository
		meta    boil.Metamap
		list    = make(boil.Metamap)
		printer = boil.NewPrinter(os.Stdout)
	)

	if repo, err = boil.OpenRepository(config.Config.GetRepositoryPath()); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if meta, err = repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}

	for k, v := range meta {
		if k = strings.ToLower(k); strings.HasPrefix(k, strings.ToLower(config.Prefix)) {
			list[k] = v
		}
	}
	if len(list) == 0 {
		printer.Printf("No templates in repository.\n")
		return nil
	}
	if config.Prefix != "" {
		printer.Printf("Templates found in current repository at %s:\n", config.Prefix)
	} else {
		printer.Printf("Templates found in current repository:\n")
	}
	printer.Printf("\n")
	list.Print(printer)

	return nil
}
