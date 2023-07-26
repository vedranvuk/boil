package exec

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is the Exec command configuration.
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

	// Vars are variables given by the user on command line.
	// These variables will be available via .Vars template field.
	Vars VarMap

	// Config is the loaded program configuration.
	Configuration *boil.Configuration
}

// ShouldPrint returns true if executin should be printed.
func (self *Config) ShouldPrint() bool {
	return self.Configuration.Overrides.Verbose || self.NoExecute
}

// GetRepositoryPath returns the RepositoryPath considering override values.
func (self *Config) GetRepositoryPath() string {
	return self.Configuration.GetRepositoryPath()
}

// state is the Exec command state.
type state struct {
	// TemplatePath is the adjusted path to the template usable by Repository.
	TemplatePath string
	// TargetDir is the adjusted absolute path to the output directory.
	TargetDir string
	// Repository is the loaded Repository.
	Repository boil.Repository
	// Metamap is the metamap of the loaded repository.
	Metamap boil.Metamap
	// Data for Template files, combined from various inputs.
	Data *Data
	// MakeBackup dictates if backups should be made on execution.
	MakeBackups bool
}

// Run executes the Exec command configured by config.
// If an error occurs it is returned and the operation may be considered as
// failed. Run modifies the passed config.
func Run(config *Config) (err error) {

	var state = &state{
		TemplatePath: config.TemplatePath,
		TargetDir:    config.TargetDir,
		MakeBackups:  config.Configuration.ShouldBackup(),
		Data:         NewData(),
	}

	if config.NoExecute {
		fmt.Printf("NoExecute enabled, printing commands instead of executing.\n")
	}

	// Open repository
	if strings.HasPrefix(config.TemplatePath, string(os.PathSeparator)) {
		// If TemplatePath is an absolute path open the Repository in the
		// template directory and adjust the template path to an empty string
		// i.e. root repository directory.
		if state.Repository, err = boil.OpenRepository(config.TemplatePath); err != nil {
			return fmt.Errorf("open repository: %w", err)
		}
		state.TemplatePath = ""
		if config.ShouldPrint() {
			fmt.Println("Absolute Template path specified, repository opened at template root.")
		}
	} else {
		if state.Repository, err = boil.OpenRepository(config.GetRepositoryPath()); err != nil {
			return fmt.Errorf("open repository: %w", err)
		}
	}
	if config.ShouldPrint() {
		fmt.Printf("Repository location: %s\n", state.Repository.Location())
	}

	// Load Metamap
	if state.Metamap, err = state.Repository.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	if config.ShouldPrint() {
		state.Metamap.Print()
	}

	// Get absolute target path and adjust it in config.
	if state.TargetDir, err = filepath.Abs(config.TargetDir); err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	// Init data.
	if err = state.Data.InitStandardVars(state); err != nil {
		return fmt.Errorf("initialize data: %w", err)
	}
	if err = state.Data.MergeVars(config.Vars); err != nil {
		return fmt.Errorf("load user variables: %w", err)
	}

	// Get a list of template file executions for this Template.
	var execs Executions
	if execs, err = PrepareExecutions(state); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return fmt.Errorf("not a boil template: %s", config.TemplatePath)
		}
		return fmt.Errorf("enumerate template files for execution: %w", err)
	}
	if config.ShouldPrint() {
		var wr = tabwriter.NewWriter(os.Stdout, 2, 2, 2, 32, 0)
		fmt.Println()
		fmt.Println("Executions:")
		fmt.Println()
		for _, exec := range execs {
			fmt.Fprintf(wr, "[Template]\t[Source]\t[Target]\n")
			for _, def := range exec.List {
				fmt.Fprintf(wr, "%s\t%s\t%s\n", exec.TemplateName, def.Source, def.Target)
			}
		}
		fmt.Fprintf(wr, "\n")
		wr.Flush()
	}

	// Add Data from Prompts.
	for _, exec := range execs {
		for _, prompt := range exec.Metafile.Prompts {
			fmt.Printf("Enter value for %s:\n", prompt.Prompt)
			var reader = bufio.NewReader(os.Stdin)
			var input string
			for {
				if input, err = reader.ReadString('\n'); err != nil {
					return fmt.Errorf("prompt input: %w", err)
				}
				state.Data.Vars[prompt.Variable] = strings.TrimSpace(input)
				break
			}
		}
	}

	// Check for existing target files if no overwrite allowed.
	if !config.Overwrite {
		if err = execs.CheckForTargetConflicts(); err != nil {
			return err
		}
	}

	// Execute and exit.
	return execs.Execute(state)
}
