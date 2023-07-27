package list

import (
	"fmt"
	"strings"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

type Config struct {
	Path string
	// Config is the loaded program configuration.
	Configuration *boil.Configuration
}

func Run(config *Config) (err error) {
	var (
		repo boil.Repository
		meta boil.Metamap
	)
	if repo, err = boil.OpenRepository(config.Configuration.GetRepositoryPath()); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if meta, err = repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	var filtered = make(boil.Metamap)
	for k, v := range meta {
		if k = strings.ToLower(k); strings.HasPrefix(k, strings.ToLower(config.Path)) {
			filtered[k] = v
		}
	}
	if config.Path != "" {
		fmt.Printf("Templates found in current repository at %s:\n", config.Path)
	} else {
		fmt.Printf("Templates found in current repository:\n")
	}
	fmt.Println()
	filtered.Print()
	return nil
}
