// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package list implements boil's list command.
package list

import (
	"fmt"
	"strings"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

type Config struct {
	// Prefix is the path prefix at which to start listing.
	Prefix string
	// Config is the loaded program configuration.
	Configuration *boil.Config
}

func Run(config *Config) (err error) {

	var (
		repo     boil.Repository
		meta     boil.Metamap
		filtered = make(boil.Metamap)
	)

	if repo, err = boil.OpenRepository(config.Configuration); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	if meta, err = repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}

	for k, v := range meta {
		if k = strings.ToLower(k); strings.HasPrefix(k, strings.ToLower(config.Prefix)) {
			filtered[k] = v
		}
	}

	if config.Prefix != "" {
		fmt.Printf("Templates found in current repository at %s:\n", config.Prefix)
	} else {
		fmt.Printf("Templates found in current repository:\n")
	}
	fmt.Println()

	filtered.Print()

	return nil
}
