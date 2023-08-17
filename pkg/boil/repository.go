// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package boil

import (
	"errors"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// OpenRepository opens a repository at the specified path. It returns an
// implementation that handles the specific path format.
//
// Currently supported backends:
// * local filesystem (DiskRepository)
//
// If an error occurs it is returned with a nil repository.
func OpenRepository(path string) (repo Repository, err error) {

	// TODO: Detect repository path and return an appropriate implementaiton.

	// TODO: Implement network loading.
	if strings.HasPrefix(strings.ToLower(path), "http") {
		return nil, errors.New("loading repositories from network not yet implemented")
	}

	// Open a directory on local fs as repository root.
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
	if path, err = filepath.Abs(path); err != nil {
		return nil, fmt.Errorf("get absolute repository path: %w", err)
	}

	return NewDiskRepository(path), nil
}

// Repository defines a location where Templates are stored.
//
// Templates inside a Repository are addressed by a path relative to the
// Repository root, i.e. 'apps/cliapp'.
type Repository interface {
	// Location returns the repository location in a format specific to
	// Repository backend.
	Location() string

	// LoadMetamap loads metadata from repository walking all child
	// subdirectories and returns it or returns a descriptive error.
	//
	// The resulting Metamap will contain a *Metadata for each subdirectory at
	// any level in the Repository that contains a Metafile, i.e. defines a
	// Template, under a key that will be a path relative to Repository.
	//
	// If the root of the Repository contains Metafile i.e. is a Template
	// itself an entry for it will be set under current directory dot ".".
	//
	// Any groups defined in a template will be added under the same path as the
	// template that defines them but with a group name suffix prefixed with a
	// "#". For example if a template 'go/app' defines a 'config' group it would
	// be addressed as 'go/app#config'.
	LoadMetamap() (Metamap, error)
	// HasMeta returns true if path contains a metafile or an error.
	HasMeta(path string) (bool, error)
	// OpenMeta returns a *Metafile at path or an error if it does not exist or
	// some other error occurs.
	OpenMeta(path string) (*Metafile, error)
	// SaveMeta saves the Metafile into repository accorsing to its Path or
	// returns an error. It overwrites the target if it exists or creates it if
	// it does not. Creates required directories along the path.
	SaveMeta(meta *Metafile) error

	// Exists returns true if file or dir exists at path or an error.
	Exists(path string) (bool, error)
	// GetFile gets contents of the file at path. It must exist and be
	// referenced in meta.
	ReadFile(path string) ([]byte, error)
	// WriteFile writes data as contents of the file at path into repository.
	// File may exists and if it does will be truncated and overwritten.
	// Path must be referenced in meta. Directories will be created as needed
	// if the file is referenced in meta.
	WriteFile(path string, data []byte) error
	// Mkdir creates all directories along the path or returns an error.
	Mkdir(path string) error
	// Remove removes file at path if path points to a file or recusrsively
	// deletes all content in the directory at path, recusrively. Returns an
	// error if one occured.
	Remove(path string) error

	// WalkDir walks the repository depth first and calls f for each file or
	// directory found in the repository. Behaves exactly like filepath.WalkDir
	// except that the path given to f will be a path relative to the repository
	// root.
	WalkDir(root string, f fs.WalkDirFunc) error
}

// IsRepoPath returns truth is the path is a path relative to repository.
// This is true if the path has no relation prefix and is not rooted.
func IsRepoPath(in string) bool {
	return !strings.HasPrefix(in, ".") && !strings.HasPrefix(in, "/")
}
