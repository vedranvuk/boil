package exec

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Execution defines an execution of a single template file.
type Execution struct {
	// Source is path of the source Template file relative to the
	// repository root to be executed. Placeholders in the Source path will be
	// expanded when determining absolute Target path.
	Source string
	// Target is the absolute path of the target file which will contain Source
	// template output. If the source path had placeholder values they will be
	// replaced with actual values.
	Target string
	// IsDir wil be true if Source is a directory.
	IsDir bool
}

// Executions defines a list of executions for a Template.
type ExecutionList struct {
	TemplateName string
	Metafile     *boil.Metafile
	List         []*Execution
}

// Executions is a list of ExecutionLists.
type Executions []*ExecutionList

// Execute executes all defined executions or returns an error.
func (self Executions) Execute(config *Config) (err error) {

	if config.Configuration.ShouldBackup() {
		var id string
		if id, err = boil.CreateBackup(config.Configuration, config.TargetDir); err != nil {
			return fmt.Errorf("create target dir backup: %w", err)
		}
		defer func() {
			if err != nil {
				if e := boil.RestoreBackup(id); e != nil {
					err = fmt.Errorf("restore backup failed after error '%w': %w", err, e)
				}
			}
		}()
	}

	for _, exec := range self {

		// Create dirs.
		for _, item := range exec.List {
			if !item.IsDir {
				continue
			}
			if err = os.MkdirAll(item.Target, os.ModePerm); err != nil {
				return fmt.Errorf("error creating target directory %s: %w", item.Target, err)
			}
		}

		// Execute source templates.
		for _, item := range exec.List {
			if item.IsDir {
				continue
			}
			var (
				buf  []byte
				tmpl *template.Template
				file *os.File
			)
			if buf, err = fs.ReadFile(config.repository, item.Source); err != nil {
				return fmt.Errorf("read template file '%s': %w", item.Source, err)
			}
			if tmpl, err = template.New(filepath.Base(item.Source)).Parse(string(buf)); err != nil {
				return fmt.Errorf("parse template file '%s': %w", item.Source, err)
			}
			if err = os.MkdirAll(filepath.Dir(item.Target), os.ModePerm); err != nil {
				return fmt.Errorf("create target file dir '%s': %w", filepath.Dir(item.Target), err)
			}
			if file, err = os.Create(item.Target); err != nil {
				return fmt.Errorf("create target file '%s': %w", item.Target, err)
			}
			defer file.Close()
			if err = tmpl.Execute(file, config.data); err != nil {
				return fmt.Errorf("execute template '%s' into target '%s': %w", item.Source, item.Target, err)
			}
		}
	}
	return nil
}

// CheckForTargetConflicts returns nil if none of the Target paths of all
// defined Executions in self do not point to an existing file. Otherwise a
// descriptive error is returned.
func (self Executions) CheckForTargetConflicts() (err error) {
	for _, execGroup := range self {
		for _, exec := range execGroup.List {
			if _, err = os.Stat(exec.Target); err != nil {
				if !errors.Is(err, os.ErrNotExist) {
					return fmt.Errorf("stat target file: %w", err)
				}
			} else {
				return fmt.Errorf("target file already exists: %s", exec.Target)
			}
		}
	}
	return nil
}

// PrepareExecutions prepares a list of executions from an Exec Config
// or returns an error if one occurs.
// It will return an error if:
// Multi with missing templates.
// Missing files or dirs.
func PrepareExecutions(config *Config) (execs Executions, err error) {
	if err = getTemplateExecutions(config, config.TemplatePath, &execs); err == nil {
		if err = validateExecutions(config, execs); err != nil {
			err = fmt.Errorf("executions validation failed: %w", err)
		}
	}
	return
}

// getTemplateExecutions uses config to recursively construct execs starting
// from path. if the function failes it returns an error.
func getTemplateExecutions(config *Config, path string, execs *Executions) (err error) {
	var meta *boil.Metafile
	var a = strings.Split(path, ":")
	if l := len(a); l < 1 || l > 2 {
		return fmt.Errorf("invalid template path '%s'", path)
	}
	path = a[0]
	if meta, err = config.metamap.Metafile(path); err != nil {
		return fmt.Errorf("load template: '%w'", err)
	}
	var list = &ExecutionList{
		TemplateName: meta.Name,
		Metafile:     meta,
	}
	for _, dir := range meta.Directories {
		list.List = append(list.List, &Execution{
			Source: filepath.Join(path, dir),
			Target: filepath.Join(config.TargetDir, config.data.ReplaceAll(dir)),
			IsDir:  true,
		})
	}
	for _, file := range meta.Files {
		if _, err := fs.Stat(config.repository, filepath.Join(path, file)); err != nil {
			return fmt.Errorf("template file '%s' stat error: %w", file, err)
		}
		list.List = append(list.List, &Execution{
			Source: filepath.Join(path, file),
			Target: filepath.Join(config.TargetDir, config.data.ReplaceAll(file)),
			IsDir:  false,
		})
	}
	*execs = append(*execs, list)
	if len(a) == 2 {
		for _, group := range meta.Groups {
			if group.Name != a[1] {
				continue
			}
			for _, name := range group.Templates {
				if err = getTemplateExecutions(config, filepath.Join(path, name), execs); err != nil {
					return
				}
			}
		}
	}
	return nil
}

// validateExecutions validates all executions.
func validateExecutions(config *Config, execs Executions) (err error) {
	for _, exec := range execs {
		if err = exec.Metafile.Validate(config.Configuration); err != nil {
			break
		}
	}
	return
}
