package boil

import (
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
func OpenRepository(config *Config) (repo Repository, err error) {
	// TODO: Detect repository path and return an appropriate implementaiton.
	var fn string
	if fn = config.Repository; config.Overrides.Repository != "" {
		fn = config.Overrides.Repository
	}
	if fn, err = filepath.Abs(fn); err != nil {
		return nil, fmt.Errorf("get absolute repo path: %w", err)
	}
	config.Runtime.LoadedRepository = fn
	return NewDiskRepository(fn), nil
}

// DiskRepository is a repository that works with an local fileystem.
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
