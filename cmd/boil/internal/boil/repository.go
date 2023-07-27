package boil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

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

// Repository defines a location where Templates are stored.
//
// Templates inside a Repository are addressed by a path relative to the
// Repository root, i.e. 'apps/cliapp'.
type Repository interface {
	// FS wraps the Open method to open a file from a repository.
	fs.FS

	// LoadMetamap loads metadata from root directory recursively recursively
	// walking all child subdirectories and returns it or returns a descriptive
	// error if one occurs.
	//
	// The resulting Metamap will contain a *Metadata for each subdirectory at
	// any level in the Repository that contains a Metafile, i.e. defines a
	// Template, under a key that will be a path relative to Repository.
	//
	// If the root of the Repository contains Metafile i.e. is a Template
	// itself an entry for it will be set under an empty key - not the standard
	//  current directory dot ".".
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
func OpenRepository(path string) (repo Repository, err error) {
	// TODO: Detect repository path and return an appropriate implementaiton.

	// See if it's http
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

	return NewDiskRepository(path), nil
}

// DiskRepository is a repository that works with a local fileystem.
// It is initialized from an absolute filesystem path or a path relative to the
// current working directory.
type DiskRepository struct {
	root string
}

// NewDiskRepository returns a new DiskRepository rooted at root.
func NewDiskRepository(root string) *DiskRepository {
	return &DiskRepository{root}
}

// Open implements the Repository.fs.FS.Open.
func (self DiskRepository) Open(name string) (fs.File, error) {
	return os.OpenFile(filepath.Join(self.root, name), os.O_RDONLY, os.ModePerm)
}

// LoadMetamap implements Repository.LoadMetamap.
func (self DiskRepository) LoadMetamap() (m Metamap, err error) {
	return MetamapFromDir(self.root)
}

// Location implements Repository.Location.
func (self DiskRepository) Location() string { return self.root }
