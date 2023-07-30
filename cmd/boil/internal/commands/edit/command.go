package edit

import (
	"fmt"
	"os/exec"
	"path/filepath"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

type Config struct {
	TemplatePath string
	// Config is the loaded program configuration.
	Configuration *boil.Configuration
}

// Run executes the Edit command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) { return newState().Run(config) }

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

// Run executes the Edit command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func (self *state) Run(config *Config) (err error) {

	if self.config = config; self.config == nil {
		return fmt.Errorf("nil config")
	}
	if self.repo, err = boil.OpenRepository(config.Configuration.GetRepositoryPath()); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if self.metamap, err = self.repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	if self.metafile, err = self.metamap.Metafile(config.TemplatePath); err != nil {
		return fmt.Errorf("template %s not found", config.TemplatePath)
	}
	
	self.vars["TemplatePath"] = filepath.Join(self.repo.Location(), config.TemplatePath)

	var args []string
	for _, arg := range config.Configuration.Editor.Arguments {
		args = append(args, self.vars.ReplaceAll(arg))
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
