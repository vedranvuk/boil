package edit

import (
	"fmt"
	"path/filepath"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is the Edit command configuration.
type Config struct {
	// TemplatePath is the path of the template to edit.
	// It may not contain group names.
	TemplatePath string
	// EditAction specifies the edit sub action.
	EditAction string
	// Config is the loaded program configuration.
	Configuration *boil.Configuration
}

// Run executes the Edit command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) { return newState().Run(config) }

// newState returns a new state.
func newState() *state {
	return &state{
		vars: make(boil.Variables),
	}
}

// state is the execution state of the edit command.
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
	if self.repo, err = boil.OpenRepository(config.Configuration); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if self.metamap, err = self.repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	if self.metafile, err = self.metamap.Metafile(config.TemplatePath); err != nil {
		return fmt.Errorf("template %s not found", config.TemplatePath)
	}

	self.vars["TemplatePath"] = filepath.Join(self.repo.Location(), config.TemplatePath)

	switch config.EditAction {
	case "edit":
		return self.config.Configuration.Editor.Execute(self.vars)
	case "all":
		err = boil.NewWizard(config.Configuration, self.metafile).EditAll()
	case "info":
		err = boil.NewWizard(config.Configuration, self.metafile).EditInfo()
	case "files":
		err = boil.NewWizard(config.Configuration, self.metafile).EditFiles()
	case "dirs":
		err = boil.NewWizard(config.Configuration, self.metafile).EditDirs()
	case "prompts":
		err = boil.NewWizard(config.Configuration, self.metafile).EditPrompts()
	case "preparse":
		err = boil.NewWizard(config.Configuration, self.metafile).EditPreParse()
	case "preexec":
		err = boil.NewWizard(config.Configuration, self.metafile).EditPreExec()
	case "postexec":
		err = boil.NewWizard(config.Configuration, self.metafile).EditPostExec()
	case "groups":
		err = boil.NewWizard(config.Configuration, self.metafile).EditGroups()
	default:
		panic("unknown edit action")
	}
	if err != nil {
		return
	}
	if config.Configuration.Overrides.Verbose {
		self.metafile.Print()
	}
	return self.repo.SaveTemplate(self.metafile)
}
