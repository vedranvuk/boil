package boil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// Repository is an interface to a repository backend.
type Repository interface {
	// FS wraps the Open method to open a file from a repository.
	fs.FS
	// LoadMetamap loads metadata from repository recursively and returns it
	// or returns a descriptive error if one occured.
	//
	// The resulting Metamap will contain a key for each subdirectory
	// recursively found in the repository. Only keys of paths to directories
	// containing a Metafile, i.e. Template directories will have a valid
	// *Metadata value. All other keys will have a nil value.
	//
	// The format of keys is a path relative to repo root i.e. 'apps/cli'.
	// The root directory is stored with a dot ".".
	LoadMetamap() (Metamap, error)
}

// OpenRepository returns an implementation of a repository depending on
// repository path from the config, respecting overrides. It sets
// config.Runtime.LoadedRepository to an absolute path of the opened repository.
//
// If an error occurs it is returned with a nil repository.
//
// Currently supported backends:
// * local filesystem (DiskRepository)
func OpenRepository(path string) (repo Repository, err error) {
	// TODO: Detect repository path and return an appropriate implementaiton.
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
	return NewDiskRepository(path), nil
}

// DiskRepository is a repository that works with  local fileystem.
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
	return LoadMetamap(self.root)
}
