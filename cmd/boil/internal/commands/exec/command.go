package exec

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

	// UserVariabled are variables given by the user on command line.
	// These variables will be available via .UserVariables template field.
	UserVariables map[string]string

	// Config is the loaded program configuration.
	Configuration *boil.Configuration

	repository        boil.Repository // loaded repository.
	metamap           boil.Metamap    // metamap of loaded repository.
	absRepositoryPath string          // loaded repository absolute path.
	data              *Data           // template file data.
}

// ShouldPrint returns true if executin should be printed.
func (self *Config) ShouldPrint() bool {
	return self.Configuration.Overrides.Verbose || self.NoExecute
}

// Run executes the Exec command configured by config.
// If an error occurs it is returned and the operation may be considered as
// failed. Run modifies the passed config.
func Run(config *Config) (err error) {

	var (
		execs Executions // list of template file executions
	)

	if config.ShouldPrint() {
		fmt.Printf("NoExecute enabled, printing commands instead of executing.\n")
	}

	// Load configuration.
	if err = config.Configuration.LoadOrCreate(); err != nil {
		return
	}
	if config.ShouldPrint() {
		fmt.Printf("Using configuration file: %s\n", config.Configuration.Runtime.LoadedConfigFile)
	}

	// Load repository.
	if strings.HasPrefix(config.TemplatePath, string(os.PathSeparator)) {
		// Absolute template path, open Template as Repository.
		config.repository, err = boil.OpenRepository(config.TemplatePath)
		config.absRepositoryPath = config.TemplatePath
		config.TemplatePath = ""
		if config.ShouldPrint() {
			fmt.Printf("Absolute template path specified, not using a repository.\n")
		}
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
		if config.ShouldPrint() {
			fmt.Printf("Using repository: %s\n", config.absRepositoryPath)
		}
	}
	if err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	// Load Metamap.
	if config.metamap, err = config.repository.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	if config.ShouldPrint() {
		var wr = tabwriter.NewWriter(os.Stdout, 2, 2, 2, 32, 0)
		var a []string
		for k := range config.metamap {
			a = append(a, k)
		}
		sort.Strings(a)
		fmt.Println()
		fmt.Println("Repository Metamap:")
		fmt.Println()
		fmt.Fprintf(wr, "[Path]\t[Parent Template Name]\n")

		for _, v := range a {
			var s = "nil"
			if config.metamap[v] != nil {
				s = config.metamap[v].Name
			}
			fmt.Fprintf(wr, "%s\t%s\n", v, s)
		}
		fmt.Fprintf(wr, "\n")
		wr.Flush()
	}

	// Get absolute target path and adjust it in config.
	if config.TargetDir, err = filepath.Abs(config.TargetDir); err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}

	// Get a list of template file executions for this Template.
	if execs, err = PrepareExecutions(config); err != nil {
		return fmt.Errorf("prepare template files for execution: %w", err)
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

	// Initialize Data.
	if err = InitConfigData(config); err != nil {
		return fmt.Errorf("prepare data: %w", err)
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
				config.data.Vars[prompt.Variable] = strings.TrimSpace(input)
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
	return execs.Execute(config)
}
