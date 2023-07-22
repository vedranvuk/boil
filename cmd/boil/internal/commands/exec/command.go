package exec

import (
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
	// Config is the loaded program configuration.
	*boil.Config
	// TemplatePath is the source template path.
	TemplatePath string
	// ModulePath is the module path to use for the project.
	ModulePath string
	// TargetDir is the target project directory.
	TargetDir string
	// NoCreateDir, if true will not create a directory based on the ProjectName
	// but instead write all files directly in the TargetDir.
	NoCreateDir bool
	// ProjectName is the custom app name to use.
	// Optional.
	ProjectName string
	// NoExecute if true does not execute any write operations.
	NoExecute bool
	// Vars are optional variables defined by the user on command invocation
	// and are passed to the template files.
	Vars map[string]string
}

// CommandNew creates a new project from a template.
func Run(cfg *Config) error {

	var (
		abs string
		err error
	)

	// Init config state.
	if err = cfg.InitializeState(); err != nil {
		return fmt.Errorf("init config state: %w", err)
	}

	// Get absolute Template path.
	if !strings.HasPrefix(cfg.TemplatePath, "/") {
		cfg.TemplatePath = filepath.Join(cfg.State.Repository, cfg.TemplatePath)
	}
	if _, err = os.Stat(cfg.TemplatePath); err != nil {
		return fmt.Errorf("template not found: %w", err)
	}

	// Parse target.
	if abs, err = filepath.Abs(cfg.TargetDir); err != nil {
		return fmt.Errorf("get absolute target path: %w", err)
	}
	cfg.TargetDir = abs

	// Load metadata.
	var meta *boil.Metadata
	if meta, err = boil.LoadMetadataFromDir(cfg.TemplatePath); err != nil {
		return err
	}

	// Get version.
	var (
		va      = strings.Split(strings.TrimPrefix(runtime.Version(), "go"), ".")
		version = va[0] + "." + va[1]
	)

	// Create data.
	var data = map[string]string{
		"ProjectName": filepath.Base(cfg.ModulePath),
		"TargetDir":   cfg.TargetDir,
		"GoVersion":   version,
		"ModulePath":  cfg.ModulePath,
	}
	if cfg.ProjectName != "" {
		data["ProjectName"] = cfg.ProjectName
	}

	var (
		dirs  = make(map[string]string)
		files = make(map[string]string)
	)

	// Create target dir names.
	for _, d := range meta.Directories {
		var (
			in  = filepath.Join(cfg.TemplatePath, d)
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
		if out, found = strings.CutPrefix(out, cfg.TemplatePath); !found {
			return fmt.Errorf("invalid source dir path: %s", in)
		}
		dirs[in] = filepath.Join(cfg.TargetDir, out)
	}

	// Create target file names.
	for _, f := range meta.Files {
		var (
			in  = filepath.Join(cfg.TemplatePath, f)
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
		if out, found = strings.CutPrefix(out, cfg.TemplatePath); !found {
			return fmt.Errorf("invalid source file path: %s", in)
		}
		files[in] = filepath.Join(cfg.TargetDir, out)
	}

	// Print some debug.
	if cfg.State.Verbose || cfg.NoExecute {
		fmt.Printf("Source template path: %s\n", cfg.TemplatePath)
		fmt.Printf("Target directory path: %s\n", cfg.TargetDir)
		fmt.Printf("Module path: %s\n", cfg.ModulePath)

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
		for k := range cfg.Vars {
			keys = append(keys, k)
		}
		sort.Strings(keys)
		for _, k := range keys {
			fmt.Printf("%s=%s\n", k, cfg.Vars[k])
		}
	}

	// Template data.
	var td = &boil.Data{
		Data: data,
		Meta: meta,
	}

	if cfg.NoExecute {
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
