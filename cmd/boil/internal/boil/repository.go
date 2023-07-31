package boil

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
)

// File is a file in a repository.
type File interface {
	fs.File
	Write(p []byte) (n int, err error)
}

// Repository defines a location where Templates are stored.
//
// Templates inside a Repository are addressed by a path relative to the
// Repository root, i.e. 'apps/cliapp'.
type Repository interface {
	// StatFS wraps the Stat method to stat a file from a repository.
	fs.StatFS

	// NewTemplate creates a new Template at the given path and returns its
	// metadata for modification and a nil error. If a template at path already
	// exists or some other error occurs it is returned.
	NewTemplate(string) (*Metafile, error)

	// SaveTemplate saves a template metafile to repository.
	SaveTemplate(*Metafile) error

	// NewDirectory creates a new directory at the specified path relative to
	// the repository root and returns nil on success or error if the target
	// directory already exists. If the path is invalid or other error occurs it
	// is returned.
	NewDirectory(string) error

	// NewFile creates a new file at the specified path relative to the
	// repository root and returns nil on success of error if the target file
	// already exists. if the path is invalid or other error occurs it is
	// returned.
	NewFile(string) (File, error)

	// OpenOrCreate opens an existing file or creates one if it doesnt exists
	// and returns it or an error if one occurs. File is open in rwr mode.
	OpenOrCreate(string) (File, error)

	// LoadMetamap loads metadata from root directory recursively recursively
	// walking all child subdirectories and returns it or returns a descriptive
	// error if one occurs.
	//
	// The resulting Metamap will contain a *Metadata for each subdirectory at
	// any level in the Repository that contains a Metafile, i.e. defines a
	// Template, under a key that will be a path relative to Repository.
	//
	// If the root of the Repository contains Metafile i.e. is a Template
	// itself an entry for it will be set under current directory dot ".".
	LoadMetamap() (Metamap, error)

	// Location returns the repository location in a format specific to
	// Repository backend.
	Location() string
}

// OpenRepository returns an implementation of a Repository depending on
// repository path format.
//
// Currently supported backends:
// * local filesystem (DiskRepository)
//
// If an error occurs it is returned with a nil repository.
func OpenRepository(config *Configuration) (repo Repository, err error) {

	var path = config.GetRepositoryPath()
	// TODO: Detect repository path and return an appropriate implementaiton.

	// TODO: Implement network loading.
	if strings.HasPrefix(strings.ToLower(path), "http") {
		return nil, errors.New("loading repositories from network not yet implemented")
	}

	// Check valid local filesystem
	var fi os.FileInfo
	if fi, err = os.Stat(path); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, fmt.Errorf("repository directory does not exist: %s", path)
		}
		return nil, fmt.Errorf("stat repository: %w", err)
	}
	if !fi.IsDir() {
		return nil, fmt.Errorf("repository path is not a directory: %s", path)
	}

	// Get absolute path
	if !strings.HasPrefix(path, string(os.PathSeparator)) {
		if path, err = filepath.Abs(path); err != nil {
			return nil, fmt.Errorf("get absolute repository path: %w", err)
		}
	}

	return NewDiskRepository(config), nil
}

// DiskRepository is a repository that works with a local fileystem.
// It is initialized from an absolute filesystem path or a path relative to the
// current working directory.
type DiskRepository struct {
	config *Configuration
}

// NewDiskRepository returns a new DiskRepository rooted at root.
func NewDiskRepository(config *Configuration) *DiskRepository {
	return &DiskRepository{config}
}

// Open implements Repository.StatFS.Open.
func (self DiskRepository) Open(name string) (file fs.File, err error) {
	if err = IsValidTemplatePath(name); err != nil {
		return
	}
	var root = self.config.GetRepositoryPath()
	var fn = filepath.Join(root, name)
	if !strings.HasPrefix(fn, root) {
		return nil, fmt.Errorf("invalid filename %s", name)
	}
	return os.OpenFile(fn, os.O_RDONLY, os.ModePerm)
}

// Stat implements Repository.StatFS.Stat.
func (self DiskRepository) Stat(name string) (file fs.FileInfo, err error) {
	if err = IsValidTemplatePath(name); err != nil {
		return
	}
	var root = self.config.GetRepositoryPath()
	var fn = filepath.Join(root, name)
	if !strings.HasPrefix(fn, root) {
		return nil, fmt.Errorf("invalid filename %s", name)
	}
	return os.Stat(fn)
}

// LoadMetamap implements Repository.LoadMetamap.
func (self DiskRepository) LoadMetamap() (m Metamap, err error) {
	return MetamapFromDir(self.config.GetRepositoryPath())
}

// Location implements Repository.Location.
func (self DiskRepository) Location() string { return self.config.GetRepositoryPath() }

// NewTemplate implements Repository.NewTemplate.
func (self DiskRepository) NewTemplate(path string) (meta *Metafile, err error) {

	if err = IsValidTemplatePath(path); err != nil {
		return
	}

	var root = self.config.GetRepositoryPath()
	var fn = filepath.Join(root, path)
	if !strings.HasPrefix(fn, root) {
		return nil, fmt.Errorf("invalid filename %s", path)
	}

	meta = NewMetafile(self.config)
	meta.Name = filepath.Base(path)

	return
}

// SaveTemplate implements
func (self DiskRepository) SaveTemplate(metafile *Metafile) (err error) {
	var buf []byte
	if buf, err = json.MarshalIndent(metafile, "", "\t"); err != nil {
		return fmt.Errorf("marshal metafile: %w", err)
	}
	if err = self.NewDirectory(metafile.Name); err != nil {
		return
	}
	var path = filepath.Join(self.config.GetRepositoryPath(), metafile.Name, MetafileName)
	if err = ioutil.WriteFile(path, buf, os.ModePerm); err != nil {
		return fmt.Errorf("write metafile: %w", err)
	}
	return nil
}

// NewDirectory implements Repository.NewDirectory.
func (self DiskRepository) NewDirectory(path string) (err error) {

	if err = IsValidTemplatePath(path); err != nil {
		return
	}

	var root = self.config.GetRepositoryPath()
	var fn = filepath.Join(root, path)
	if !strings.HasPrefix(fn, root) {
		return fmt.Errorf("invalid filename %s", path)
	}

	return os.MkdirAll(fn, os.ModePerm)
}

// NewFile implements Repository.NewFile.
func (self DiskRepository) NewFile(path string) (file File, err error) {

	if err = IsValidTemplatePath(path); err != nil {
		return
	}

	var root = self.config.GetRepositoryPath()
	var fn = filepath.Join(root, path)
	if !strings.HasPrefix(fn, root) {
		return nil, fmt.Errorf("invalid filename %s", path)
	}

	return os.OpenFile(path, os.O_CREATE|os.O_EXCL|os.O_RDWR, os.ModePerm)
}

func (self DiskRepository) OpenOrCreate(path string) (file File, err error) {
	if err = IsValidTemplatePath(path); err != nil {
		return
	}

	var root = self.config.GetRepositoryPath()
	var fn = filepath.Join(root, path)
	if !strings.HasPrefix(fn, root) {
		return nil, fmt.Errorf("invalid filename %s", path)
	}

	return os.OpenFile(path, os.O_CREATE|os.O_RDWR, os.ModePerm)
}

// IsValidTemplatePath returns an error if the path is of invalid format.
func IsValidTemplatePath(path string) (err error) {
	if TemplatePathIsAbsolute(path) {
		return fmt.Errorf("invalid template path: must be a path relative to repository.")
	}
	if strings.Contains(path, "#") {
		return fmt.Errorf("invalid template path: must not name a group.")
	}
	return nil
}

// ParseTemplatePath parses the input string representing a Template path into
// path and group parts which are delimited with first occurence of a hashtag.
//
// If the input is an empty string a "." path is returned and an empty group.
func ParseTemplatePath(input string) (path, group string) {
	if input == "" {
		return ".", ""
	}
	path, group, _ = strings.Cut(input, "#")
	return
}

// TemplatePathIsAbsolute returns true if the path begins with a path separator.
func TemplatePathIsAbsolute(path string) bool {
	return strings.HasPrefix(path, string(os.PathSeparator))
}
