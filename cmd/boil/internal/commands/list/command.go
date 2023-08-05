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
		repo boil.Repository
		meta boil.Metamap
		list = make(boil.Metamap)
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
	if config.Prefix != "" {
		fmt.Printf("Templates found in current repository at %s:\n", config.Prefix)
	} else {
		fmt.Printf("Templates found in current repository:\n")
	}
	fmt.Println()
	list.Print()

	return nil
}
