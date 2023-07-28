package snap

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is the SNap command configuration.
type Config struct {
	// TemplatePath is the path under which the Template will be stored
	// relative to the loaded repository root.
	TemplatePath string

	// SourcePath is an optional path to the source directory or file.
	// If ommitted a snapshot of the current directory is created.
	SourcePath string

	// Wizard specifies if a template wizard should be used.
	Wizard bool

	// Force overwriting template if it already exists.
	Overwrite bool

	// Configuration is the loaded program configuration.
	Configuration *boil.Configuration
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
	src      string
	rootinfo fs.FileInfo
	files    []string
	dirs     []string
}

func (self *state) Run(config *Config) (err error) {

	if self.config = config; self.config == nil {
		return fmt.Errorf("nil config")
	}

	if err = boil.IsValidTemplatePath(config.TemplatePath); err != nil {
		return err
	}

	if self.repo, err = boil.OpenRepository(config.Configuration.GetRepositoryPath()); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	if self.metamap, err = self.repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}

	if _, err = self.metamap.Metafile(config.TemplatePath); err == nil && !config.Overwrite {
		return fmt.Errorf("template %s already exists", config.TemplatePath)
	}

	if self.src, err = filepath.Abs(config.SourcePath); err != nil {
		return fmt.Errorf("get absolute source path: %w", err)
	}

	if self.rootinfo, err = os.Stat(self.src); err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	if self.rootinfo.IsDir() {
		if err = filepath.WalkDir(self.src, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				self.dirs = append(self.dirs, path)
			} else {
				self.files = append(self.files, path)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("enumerate source directory: %w", err)
		}
	} else {
		self.files = append(self.files, self.src)
	}
	if self.config.Configuration.Overrides.Verbose {
		fmt.Printf("Source directories:\n")
		for _, dir := range self.dirs {
			fmt.Printf("%s\n", dir)
		}
		fmt.Println()
		fmt.Printf("Source files:\n")
		for _, file := range self.files {
			fmt.Printf("%s\n", file)
		}
		fmt.Println()
	}

	if self.metafile, err = self.repo.NewTemplate(config.TemplatePath); err != nil {
		return fmt.Errorf("create new template: %w", err)
	}
	self.metafile.Name = filepath.Base(config.TemplatePath)

	if config.Wizard {
		if err = NewWizard(self).Execute(); err != nil {
			return fmt.Errorf("execute wizard: %w", err)
		}
		if err = self.metafile.Save(); err != nil {
			return
		}
	}

	if !config.Overwrite {
		for _, file := range self.files {
			if _, err = self.repo.Stat(file); err == nil {
				return fmt.Errorf("template file '%s' already exists", file)
			} else {
				if !errors.Is(err, fs.ErrNotExist) {
					return fmt.Errorf("stat template file '%s': %w", file, err)
				}
			}
		}
	}

	for _, file := range self.files {
		_ = file
	}

	return
}
