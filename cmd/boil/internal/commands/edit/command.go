package edit

import (
	"fmt"

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
	_ = repo
	_ = meta
	return nil
}
