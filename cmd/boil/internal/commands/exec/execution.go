package exec

import (
	"bufio"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"text/tabwriter"
	"text/template"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Execution defines the source and target of a template file to be executed.
type Execution struct {
	// Source is unmodified path of the source Template file relative to the
	// loaded repository root to be executed.
	Source string
	// Target is the absolute path of the target file which will contain Source
	// template output. If the source path had placeholder values they will be
	// replaced with actual values to generate output filenames.
	Target string
	// IsDir wil be true if Source is a directory.
	IsDir bool
}

// Template defines a list of template files to execute for a Template.
type Template struct {
	// Name is the name of the template.
	Name string
	// Metafile is the Template Metafile.
	Metafile *boil.Metafile
	// List is a list of executions to be performed as for this template.
	List []*Execution
}

// Templates is a list of Template.
// It holds a list of groups of files to execute for a Template.
type Templates []*Template

// GetTemplatesForState returns Templates to be executed from a state. It
// returns empty Templates and an error if the state is invalid, one or more
// template files is missing, any group addresses a missing template or some
// other error.
func GetTemplatesForState(state *state) (templates Templates, err error) {
	err = produceTemplates(state, state.TemplatePath, &templates)
	return
}

// produceTemplates uses state to recursively construct execs starting
// from path. if the function failes it returns an error.
func produceTemplates(state *state, path string, templates *Templates) (err error) {

	var (
		meta  *boil.Metafile
		group string
	)

	path, group = boil.ParseTemplatePath(path)

	if meta, err = state.Metamap.Metafile(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return err
		}
		return fmt.Errorf("load template: '%w'", err)
	}

	var list = &Template{
		Name:     meta.Name,
		Metafile: meta,
	}

	for _, dir := range meta.Directories {
		list.List = append(list.List, &Execution{
			Source: filepath.Join(path, dir),
			Target: filepath.Join(state.TargetDir, dir),
			IsDir:  true,
		})
	}

	for _, file := range meta.Files {
		if _, err := fs.Stat(state.Repository, filepath.Join(path, file)); err != nil {
			return fmt.Errorf("stat template file '%s': %w", file, err)
		}
		list.List = append(list.List, &Execution{
			Source: filepath.Join(path, file),
			Target: filepath.Join(state.TargetDir, file),
			IsDir:  false,
		})
	}

	*templates = append(*templates, list)

	if group != "" {
		for _, g := range meta.Groups {
			if g.Name == group {
				continue
			}
			for _, name := range g.Templates {
				if err = produceTemplates(state, filepath.Join(path, name), templates); err != nil {
					return
				}
			}
		}
	}

	return nil
}

// PromptAnswered is a callback called by Templates.PresentPrompts for each
// prompt answered by the user where variable is the name of the variable the
// prompt was for and the value is the value enetered. If the function returns
// an error it is propagated back to the caller.
type PromptAnswered func(variable, value string) error

// PresentPrompts presents a prompt to the user on command line for each of
// the prompts defined in all Templates in self, in order as they appear in
// self, depth first.
//
// Method calls callback for each prompt and returns nil if all prompts were
// successfully answered. If not, an error that the first answered callback
// returned is returned. See PromptAnswered for more.
func (self Templates) PresentPrompts(callback PromptAnswered) (err error) {
	for _, template := range self {
		for _, prompt := range template.Metafile.Prompts {
			fmt.Printf("Enter value for %s:\n", prompt.Prompt)
			var reader = bufio.NewReader(os.Stdin)
			var input string
			for {
				if input, err = reader.ReadString('\n'); err != nil {
					return fmt.Errorf("prompt input: %w", err)
				}
				if err = callback(prompt.Variable, strings.TrimSpace(input)); err != nil {
					return
				}
				break
			}
		}
	}
	return nil
}

// ExpandExecutionTargets expands all Execution.Target values of all Templates
// in self using data and returns nil. If an error occurs it is returned and
// self may be considered invalid in undetermined state.
func (self Templates) ExpandExecutionTargets(data *Data) (err error) {
	for _, template := range self {
		for _, execution := range template.List {
			execution.Target = data.ReplaceAll(execution.Target)
		}
	}
	return
}

// CheckForTargetConflicts returns nil if none of the Target paths of all
// defined Executions in self do not point to an existing file. Otherwise a
// descriptive error is returned.
func (self Templates) CheckForTargetConflicts() (err error) {
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

// Validate validates self.
func (self Templates) Validate(state *state) (err error) {
	for _, template := range self {
		if err = template.Metafile.Validate(state.Repository); err != nil {
			break
		}
	}
	return
}

// Execute executes all defined executions or returns an error.
func (self Templates) Execute(state *state) (err error) {

	if state.MakeBackups {
		var id string
		if id, err = boil.CreateBackup(state.TargetDir); err != nil {
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
			if buf, err = fs.ReadFile(state.Repository, item.Source); err != nil {
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
			if err = tmpl.Execute(file, state.Data); err != nil {
				return fmt.Errorf("execute template '%s' into target '%s': %w", item.Source, item.Target, err)
			}
		}
	}
	return nil
}

// Print prints self to stdout.
func (self Templates) Print() {
	var wr = tabwriter.NewWriter(os.Stdout, 2, 2, 2, 32, 0)
	fmt.Println()
	fmt.Println("Executions:")
	fmt.Println()
	for _, exec := range self {
		fmt.Fprintf(wr, "[Template]\t[Source]\t[Target]\n")
		for _, def := range exec.List {
			fmt.Fprintf(wr, "%s\t%s\t%s\n", exec.Name, def.Source, def.Target)
		}
	}
	fmt.Fprintf(wr, "\n")
	wr.Flush()
}
