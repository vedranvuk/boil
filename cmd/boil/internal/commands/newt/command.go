package newt

import (
	"fmt"
	"os/exec"
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
	self.metafile.Author.Name = self.config.Configuration.DefaultAuthor.Name
	self.metafile.Author.Email = self.config.Configuration.DefaultAuthor.Email
	self.metafile.Author.Homepage = self.config.Configuration.DefaultAuthor.Homepage
	if err = boil.NewWizard(self.config.Configuration, self.metafile).Execute(); err != nil {
		return fmt.Errorf("execute wizard: %w", err)
	}
	if err = self.repo.SaveTemplate(self.metafile); err != nil {
		return
	}
	self.vars["TemplatePath"] = filepath.Join(self.repo.Location(), config.TemplatePath)

	var args []string
	for _, arg := range config.Configuration.Editor.Arguments {
		args = append(args, self.vars.ReplacePlaceholders(arg))
	}
	cmd := exec.Command(config.Configuration.Editor.Program, args...)
	var buf []byte
	buf, err = cmd.Output()
	fmt.Print(string(buf))
	if err != nil {
		return fmt.Errorf("exec editor on '%s': %w", cmd.Path, err)
	}
	return nil
}
