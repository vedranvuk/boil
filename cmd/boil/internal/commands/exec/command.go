package exec

import (
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"text/template"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is a Create command configuration.
type Config struct {
	// TemplatePath is the source template path. During Run() it is adjusted to
	// an absolute Template path relative to working directory.s
	//
	// If the path is rooted, i.e. starts with "/" the path is treated as
	// absolute path to a Template and no repository is being loaded or used.
	//
	// If the path is not rooted, the path is treated as a path to a Template
	// relative to the loaded repository.
	TemplatePath string

	// TargetDir is the output directory where Template will be executed.
	// See NoTargetDir description for more details.
	TargetDir string

	// ProjectName is the custom app name to use.
	// Optional.
	ProjectName string

	// NoCreateDir, if true will not create a directory for the Template in the
	// TargetDir but instead write all files directly to the TargetDir.
	//
	// If false, a new directory will be created in the TargetDir for Template
	// output. It's name will be generated in the following way, priority first:
	// 1. If not empty, from ProjectName.
	// 2. From the last path element of the ModulePath.
	NoCreateDir bool

	// ModulePath is the module path to use when generating go.mod files.
	ModulePath string

	// NoExecute if true does not execute any write operations.
	NoExecute bool
	// Overwrite specifies wether to overwrite any existing files in the target directory.
	Overwrite bool
	// UserVariabled are variables given by the user on command line.
	UserVariables map[string]string
	// Config is the loaded program configuration.
	*boil.Config
}

// CommandNew creates a new project from a template.
func Run(config *Config) (err error) {

	var (
		repo    boil.Repository // loaded repository
		meta    *boil.Metadata  // metadata of selected Template
		metamap boil.Metamap    // metamap of the repository
		execs   Executions      // list of template file executions
	)

	// Config
	if err = config.LoadOrCreate(); err != nil {
		return
	}
	// Repository
	if repo, err = boil.OpenRepository(config.Config); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	// Metamap
	if metamap, err = repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}
	_ = metamap
	// Get absolute Template path and adjust it in config.
	if !strings.HasPrefix(config.TemplatePath, "/") {
		config.TemplatePath = filepath.Join(config.Runtime.LoadedRepository, config.TemplatePath)
	}
	if _, err = os.Stat(config.TemplatePath); err != nil {
		return fmt.Errorf("template not found: %w", err)
	}
	// Load metadata.
	if meta, err = boil.LoadMetadataFromDir(config.TemplatePath); err != nil {
		return err
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
		for _, exec := range execs {
			if _, err = os.Stat(exec.Target); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("stat target file: %w", err)
				}
			} else {
				return fmt.Errorf("target file already exists: %s", exec.Target)
			}
		}
	}
	// Get version.
	var (
		va      = strings.Split(strings.TrimPrefix(runtime.Version(), "go"), ".")
		version = va[0] + "." + va[1]
	)

	// Create data.
	var data = map[string]string{
		"ProjectName": filepath.Base(config.ModulePath),
		"TargetDir":   config.TargetDir,
		"GoVersion":   version,
		"ModulePath":  config.ModulePath,
	}
	if config.ProjectName != "" {
		data["ProjectName"] = config.ProjectName
	}

	var (
		dirs  = make(map[string]string)
		files = make(map[string]string)
	)

	// Create target dir names.
	for _, d := range meta.Directories {
		var (
			in  = filepath.Join(config.TemplatePath, d)
			out string
		)
		if _, err := os.Stat(in); err != nil {
			return fmt.Errorf("stat error on template directory %s: %w", in, err)
		}
		out = in
		for k, v := range data {
			out = strings.ReplaceAll(out, "$"+k, v)
		}
		var found bool
		if out, found = strings.CutPrefix(out, config.TemplatePath); !found {
			return fmt.Errorf("invalid source dir path: %s", in)
		}
		dirs[in] = filepath.Join(config.TargetDir, out)
	}

	// Create target file names.
	for _, f := range meta.Files {
		var (
			in  = filepath.Join(config.TemplatePath, f)
			out string
		)
		if _, err := os.Stat(in); err != nil {
			return fmt.Errorf("stat error on template file %s: %w", in, err)
		}
		out = in
		for k, v := range data {
			out = strings.ReplaceAll(out, "$"+k, v)
		}
		var found bool
		if out, found = strings.CutPrefix(out, config.TemplatePath); !found {
			return fmt.Errorf("invalid source file path: %s", in)
		}
		files[in] = filepath.Join(config.TargetDir, out)
	}

	// Print some debug.
	if config.Overrides.Verbose || config.NoExecute {
		fmt.Printf("Source template path: %s\n", config.TemplatePath)
		fmt.Printf("Target directory path: %s\n", config.TargetDir)
		fmt.Printf("Module path: %s\n", config.ModulePath)

		fmt.Printf("Target directories:\n")
		for _, v := range dirs {
			fmt.Printf("%s\n", v)
		}
		fmt.Printf("Target files:\n")
		for _, v := range files {
			fmt.Printf("%s\n", v)
		}
		fmt.Printf("Variables:\n")
		var keys []string
		for k := range config.UserVariables {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s=%s\n", k, config.UserVariables[k])
		}
	}

	// Template data.
	var td = &boil.Data{
		Data: data,
		Meta: meta,
	}

	if config.NoExecute {
		fmt.Println("no-execute specified, no write operations will be executed.")
		return nil
	}

	// Create dirs.
	for _, tgt := range dirs {
		if err := os.MkdirAll(tgt, os.ModePerm); err != nil {
			return fmt.Errorf("error creating target directory %s: %w", tgt, err)
		}
	}

	// Execute source templates.
	for src, tgt := range files {
		var (
			err  error
			buf  []byte
			tmpl *template.Template
			file *os.File
		)
		if buf, err = ioutil.ReadFile(src); err != nil {
			return fmt.Errorf("error reading source dir %s: %w", src, err)
		}
		if tmpl, err = template.New(filepath.Base(src)).Parse(string(buf)); err != nil {
			return fmt.Errorf("error parsing source template %s: %w", src, err)
		}
		if file, err = os.Create(tgt); err != nil {
			return fmt.Errorf("error creating target file %s: %w", tgt, err)
		}
		defer file.Close()
		if err = tmpl.Execute(file, td); err != nil {
			return fmt.Errorf("error executing template %s: %w", tgt, err)
		}
	}

	return nil
}
