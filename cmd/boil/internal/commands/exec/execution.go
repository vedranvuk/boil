package exec

import (
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Execution defines an execution of a single template file.
type Execution struct {
	// Source is the absolute path of the source file which is to be executed
	// onto Target path.
	// The path may be contain placeholder values.
	Source string
	// Target is the absolute path of the target file which will contain Source
	// template output. If the source path had placeholder values they will be
	// replaced with actual values.
	Target string
	// IsDir wil be true if Source is a directory.
	IsDir bool
}

// Execute executes a FileCopy operation or returns an error.
func (self *Execution) Execute(data interface{}) error {

	var (
		err  error
		buf  []byte
		tmpl *template.Template
		file *os.File
	)

	if buf, err = ioutil.ReadFile(self.Source); err != nil {
		return fmt.Errorf("read source template '%s': %w", self.Source, err)
	}

	if tmpl, err = template.New(filepath.Base(self.Source)).Parse(string(buf)); err != nil {
		return fmt.Errorf("parse source template '%s': %w", self.Source, err)
	}

	if file, err = os.Create(self.Target); err != nil {
		return fmt.Errorf("create target file '%s': %w", self.Target, err)
	}
	defer file.Close()

	if err = tmpl.Execute(file, data); err != nil {
		return fmt.Errorf("execute template '%s' to '%s': %w", self.Source, self.Target, err)
	}

	return nil
}

// TemplateExecution defines a list of executions for a Template.
type TemplateExecution struct {
	TemplateName string
	Executions
}

type TemplateExecutions []*TemplateExecution

// Execute executes all defined executions or returns an error.
func (self TemplateExecutions) Execute(config *Config) error {

	if config.NoExecute {
		fmt.Println("no-execute specified, no write operations will be executed.")
		return nil
	}

	// No-Execute.
	for _, exec := range self {

		// Print instead of executing.
		if config.NoExecute {
			fmt.Printf("Listing write operations for template '%s':\n", exec.TemplateName)
			for _, e := range exec.Executions {
				fmt.Printf("Execute '%s' to '%s' (directory: %t)\n", e.Source, e.Target, e.IsDir)
			}
			fmt.Println()
			return nil
		}

		// TODO: Backup target dir and restore on error.

		// Create dirs.
		for _, e := range exec.Executions {
			if !e.IsDir {
				continue
			}
			if err := os.MkdirAll(e.Target, os.ModePerm); err != nil {
				return fmt.Errorf("error creating target directory %s: %w", e.Target, err)
			}
		}

		// Execute source templates.
		for _, e := range exec.Executions {
			if e.IsDir {
				continue
			}
			var (
				err  error
				buf  []byte
				tmpl *template.Template
				file *os.File
				src  = e.Source
				tgt  = e.Target
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
			if err = tmpl.Execute(file, config.data); err != nil {
				return fmt.Errorf("error executing template %s: %w", tgt, err)
			}
		}
	}
	return nil
}

// Executions is a list of executions needed to execute a Boil template.
type Executions []*Execution

// PrepareExecutions prepares a list of executions from an Exec Config
// or returns an error if one occurs.
// It will return an error if:
// Multi with missing templates.
// Missing files or dirs.
func PrepareExecutions(config *Config) (execs TemplateExecutions, err error) {
	err = getTemplateExecutions(config, config.TemplatePath, execs)
	return
}

func getTemplateExecutions(config *Config, path string, execs TemplateExecutions) (err error) {
	var meta *boil.Metafile
	if meta, err = config.metamap.Metadata(path); err != nil {
		return fmt.Errorf("load template: '%w'", err)
	}
	var texec = &TemplateExecution{
		TemplateName: meta.Name,
	}
	// Create target dir names.
	for _, d := range meta.Directories {
		var (
			in  = d
			out string
		)
		if _, err := fs.Stat(config.repository, in); err != nil {
			return fmt.Errorf("stat error on template directory %s: %w", in, err)
		}
		out = config.data.ReplaceAll(in)
		var found bool
		if out, found = strings.CutPrefix(out, config.absRepositoryPath); !found {
			return fmt.Errorf("invalid source dir path: %s", in)
		}
		texec.Executions = append(texec.Executions, &Execution{
			Source: in,
			Target: filepath.Join(config.TargetDir, out),
			IsDir:  true,
		})
	}
	// Create target file names.
	for _, f := range meta.Files {
		var (
			in  = f
			out string
		)
		if _, err := os.Stat(in); err != nil {
			return fmt.Errorf("stat error on template file %s: %w", in, err)
		}
		out = config.data.ReplaceAll(in)
		var found bool
		if out, found = strings.CutPrefix(out, config.absRepositoryPath); !found {
			return fmt.Errorf("invalid source file path: %s", in)
		}
		texec.Executions = append(texec.Executions, &Execution{
			Source: in,
			Target: filepath.Join(config.TargetDir, out),
			IsDir:  false,
		})
	}
	// Create executions for multi.
	for _, m := range meta.Groups {
		for _, t := range m.Templates {
			if err = getTemplateExecutions(config, filepath.Join(path, t), execs); err != nil {
				return
			}
		}
	}

	return nil
}
