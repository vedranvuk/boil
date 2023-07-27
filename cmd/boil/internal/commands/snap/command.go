package snap

import (
	"bufio"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

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
func Run(config *Config) (err error) {

	var (
		repo  boil.Repository
		meta  boil.Metamap
		data  *boil.Metafile
		src   string
		finfo fs.FileInfo
		files []string
		dirs  []string
	)

	if err = boil.IsValidTemplatePath(config.TemplatePath); err != nil {
		return err
	}

	if repo, err = boil.OpenRepository(config.Configuration.GetRepositoryPath()); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}

	if meta, err = repo.LoadMetamap(); err != nil {
		return fmt.Errorf("load metamap: %w", err)
	}

	if _, err = meta.Metafile(config.TemplatePath); err == nil && !config.Overwrite {
		return fmt.Errorf("template %s already exists", config.TemplatePath)
	}

	if src, err = filepath.Abs(config.SourcePath); err != nil {
		return fmt.Errorf("get absolute source path: %w", err)
	}

	if finfo, err = os.Stat(src); err != nil {
		return fmt.Errorf("stat source: %w", err)
	}

	if finfo.IsDir() {
		if err = filepath.WalkDir(src, func(path string, d fs.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if d.IsDir() {
				dirs = append(dirs, path)
			} else {
				files = append(files, path)
			}
			return nil
		}); err != nil {
			return fmt.Errorf("enumerate source directory: %w", err)
		}
	} else {
		files = append(files, src)
	}

	if data, err = repo.NewTemplate(config.TemplatePath); err != nil {
		return fmt.Errorf("create new template: %w", err)
	}
	data.Name = filepath.Base(config.TemplatePath)

	return MetafileWizard(data)
}

// MetafileWizard prompts user to fill Metafile values.
func MetafileWizard(file *boil.Metafile) error {

	file.Description = promptValue("Enter template description:", ".*")
	file.Author.Name = promptValue("Enter template author name:", ".*")

	return file.Save()
}

func promptValue(prompt, regex string) string {
	fmt.Printf("%s:\n", prompt)
	var reader = bufio.NewReader(os.Stdin)
	var input string
	var err error
	for {
		if input, err = reader.ReadString('\n'); err != nil {
			fmt.Println(err)
			return ""
		}
		return strings.TrimSpace(input)
	}
}
