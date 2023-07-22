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

// MetafileName is the filename of the metafile that describes a boil template.
const MetafileName = "boil.json"

// Multi defines an execution of multiple templates at once possibly in
// the same target directory.
type Multi struct {
	// Name is the multi name.
	Name string `json:"name,omitempty"`
	// Description is the description.
	Description string `json:"description,omitempty"`
	// Templates is the array of names of templates to execute as this Multi,
	// in the order to be executed.
	Templates []string `json:"templates,omitempty"`
}

// Metadata is the bojler template metadata.
// It is found in a template dir or a multi dir that contains multiple template
// subdirectories.
type Metadata struct {
	// Name is the template name.
	Name string `json:"name,omitempty"`
	// Description is the template description.
	Description string `json:"description,omitempty"`
	// Author are the template author details.
	Author *Author `json:"author,omitempty"`
	// Version is the template version.
	Version string `json:"version,omitempty"`
	// URL is the cannonical template url.
	URL string `json:"url,omitempty"`
	// Files is a list of files contained by the template.
	Files []string `json:"files,omitempty"`
	// Directories is a list of directories contained by the template.
	Directories []string `json:"directories,omitempty"`
	// Multis is a slice of Multi-templates.
	Multis []*Multi `json:"multis,omitempty"`
	// Actions are actions to execute with the template.
	Actions struct {
		// Pre-execution commands.
		Pre []*Command `json:"pre,omitempty"`
		// Post execution commands.
		Post []*Command `json:"post,omitempty"`
	} `json:"actions,omitempty"`

	// directory is the directory from which metadata was loaded from.
	directory string
}

// ErrNoMetadata is returned by LoadMetadataFromDir if a metadata file
// does not exist in specified directory.
var ErrNoMetadata = errors.New("no metadata found")

// LoadMetadataFromDir loads metadata from dir and returns it or an error.
func LoadMetadataFromDir(dir string) (metadata *Metadata, err error) {
	var buf []byte
	if buf, err = ioutil.ReadFile(filepath.Join(dir, MetafileName)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, ErrNoMetadata
		}
		return nil, errors.New("not a boil template directory, missing metafile")
	}
	metadata = &Metadata{directory: dir}
	if err = json.Unmarshal(buf, metadata); err != nil {
		return nil, fmt.Errorf("parse metafile: %w", err)
	}
	return
}

// Metamap maps a template path to metadata loaded from its directory.
type Metamap map[string]*Metadata

// LoadMetamap loads metadata from root directory recursively and returns it
// or an error.
//
// The resulting Metamap will contain a key for each subdirectory found in root
// i.e. 'apps/webapp'.
//
// Directories that do not contain a boil metafile will have a nil Metadata.
func LoadMetamap(root string) (m Metamap, err error) {
	var d *Metadata
	m = make(Metamap)
	err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		d = nil
		if d, err = LoadMetadataFromDir(path); err != nil {
			if !errors.Is(err, ErrNoMetadata) {
				return fmt.Errorf("load metamap: %w", err)
			}
			return nil
		}
		var s = strings.TrimPrefix(strings.TrimPrefix(path, root), string(os.PathSeparator))
		m[s] = d
		return nil
	})
	return
}

// WithMetadata returns a subset of self with only the keys that have
// a non-nil metadata.
func (self Metamap) WithMetadata() (m Metamap) {
	m = make(Metamap)
	for k, v := range self {
		if v != nil {
			m[k] = v
		}
	}
	return
}
