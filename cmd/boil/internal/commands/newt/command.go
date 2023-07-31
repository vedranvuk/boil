package newt

import (
	"fmt"
	"path/filepath"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

type Config struct {
	TemplatePath  string
	Overwrite     bool
	Configuration *boil.Configuration
}

func Run(config *Config) error { return newState().Run(config) }

func newState() *state {
	return &state{
		vars: make(boil.Variables),
	}
}

type state struct {
	config   *Config
	repo     boil.Repository
	metamap  boil.Metamap
	metafile *boil.Metafile
	vars     boil.Variables
}

func (self *state) Run(config *Config) (err error) {
	if self.config = config; self.config == nil {
		return fmt.Errorf("nil config")
	}
	if err = boil.IsValidTemplatePath(config.TemplatePath); err != nil {
		return err
	}
	if self.repo, err = boil.OpenRepository(config.Configuration); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if self.metamap, err = self.repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	if _, err = self.metamap.Metafile(config.TemplatePath); err == nil && !config.Overwrite {
		return fmt.Errorf("template %s already exists", config.TemplatePath)
	}
	if self.metafile, err = self.repo.NewTemplate(config.TemplatePath); err != nil {
		return fmt.Errorf("create new template: %w", err)
	}
	if err = boil.NewWizard(self.config.Configuration, self.metafile).Execute(); err != nil {
		return fmt.Errorf("execute wizard: %w", err)
	}
	if err = self.repo.SaveTemplate(self.metafile); err != nil {
		return
	}
	self.vars["TemplatePath"] = filepath.Join(self.repo.Location(), config.TemplatePath)
	return self.config.Configuration.Editor.Execute(self.vars)
}
