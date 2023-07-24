package exec

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is a Create command configuration.
type Config struct {
	// TemplatePath is the source template path. During Run() it is adjusted to
	// an absolute path to the Template either inside or outside of repository.
	//
	// If the path is rooted, i.e. starts with "/" the path is treated as an
	// absolute path to a Template and no repository is being loaded or used.
	//
	// If the path is not rooted, the path is treated as a path to a Template
	// relative to the loaded repository.
	//
	// If TemplatePath is an absolute filesystem path it is adjusted to an
	// empty string during Run().
	TemplatePath string

	// TargetDir is the output directory where Template will be executed.
	// If the value is empty the Template will be executed in the current
	// working directory
	//
	// TargetPath is adjusted to an absolute path of TargetDir during Run().
	TargetDir string

	// NoExecute, if true will not execute any write operations and will
	// instead print out the operations like boil.Config.Verbose was enabled.
	NoExecute bool

	// Overwrite, if true specifies that any file matching a Template output
	// file already existing in the target directory may be overwritten without
	// prompting the user or generating an error.
	Overwrite bool

	// UserVariabled are variables given by the user on command line.
	// These variables will be available via .UserVariables template field.
	UserVariables map[string]string

	// Config is the loaded program configuration.
	Configuration *boil.Configuration

	repository        boil.Repository // loaded repository.
	metamap           boil.Metamap    // metamap of loaded repository.
	absRepositoryPath string          // loaded repository absolute path.
	data              *Data
}

// Run executes the Exec command configured by config.
// If an error occurs it is returned and the operation may be considered as
// failed. Run modifies the passed config.
func Run(config *Config) (err error) {

	var (
		execs TemplateExecutions // list of template file executions
	)
	// Config
	if err = config.Configuration.LoadOrCreate(); err != nil {
		return
	}
	// Repository
	if strings.HasPrefix(config.TemplatePath, string(os.PathSeparator)) {
		// Absolute template path, open Template as Repository.
		config.repository, err = boil.OpenRepository(config.TemplatePath)
		config.absRepositoryPath = config.TemplatePath
		config.TemplatePath = ""
	} else {
		var path string
		if path = config.Configuration.Repository; config.Configuration.Overrides.Repository != "" {
			path = config.Configuration.Overrides.Repository
		}
		if path, err = filepath.Abs(path); err != nil {
			return fmt.Errorf("get absolute repo path: %w", err)
		}
		config.absRepositoryPath = path
		config.repository, err = boil.OpenRepository(path)
	}
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	// Metamap
	if config.metamap, err = config.repository.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	if config.Configuration.Overrides.Verbose {
		var a []string
		for k := range config.metamap {
			a = append(a, k)
		}
		sort.Strings(a)
		fmt.Printf("Metamap:\n")
		for _, v := range a {
			var s = "nil"
			if config.metamap[v] != nil {
				s = config.metamap[v].Name
			}
			fmt.Printf("%s\t%s\n", v, s)
		}
		fmt.Println()
	}
	// Get absolute target path and adjust it in config.
	if config.TargetDir, err = filepath.Abs(config.TargetDir); err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}
	// Get a list of template file executions for this Template.
	if execs, err = PrepareExecutions(config); err != nil {
		return fmt.Errorf("prepare template files for execution: %w", err)
	}
	// Check for existing target files if no overwrite allowed.
	if !config.Overwrite {
		for _, execGroup := range execs {
			for _, exec := range execGroup.Executions {
				if _, err = os.Stat(exec.Target); err != nil {
					if !errors.Is(err, os.ErrNotExist) {
						return fmt.Errorf("stat target file: %w", err)
					}
				} else {
					return fmt.Errorf("target file already exists: %s", exec.Target)
				}
			}
		}
	}
	// Get version.
	var (
		va      = strings.Split(strings.TrimPrefix(runtime.Version(), "go"), ".")
		version = va[0] + "." + va[1]
		_       = version
	)
	// TODO Create Data.
	// Execute and exit.
	return execs.Execute(config)
}
