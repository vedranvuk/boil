package info

import (
	"fmt"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

type Config struct {
	TemplatePath string
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
}

// Run executes the Snap command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func (self *state) Run(config *Config) (err error) {

	// Checks
	if self.config = config; self.config == nil {
		return fmt.Errorf("nil config")
	}
	if err = boil.IsValidTemplatePath(config.TemplatePath); err != nil {
		return err
	}

	// Open repository and get its metamap, check template exists.
	if self.repo, err = boil.OpenRepository(config.Configuration); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if self.metamap, err = self.repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	if self.metafile, err = self.metamap.Metafile(config.TemplatePath); err != nil {
		return fmt.Errorf("template %s not found", config.TemplatePath)
	}

	self.metafile.Print()

	return nil
}
