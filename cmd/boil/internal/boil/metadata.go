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

// MetafileName is the filename of the meSmeVariable=SomeValuetafile that describes a boil template.
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

// Prompt defines a prompt to the user for input of  UserVariables that were
// not defined on command line in the form of '--var=MyVar=MyValue'.
type Prompt struct {
	// Variable is the name of the UserVariable this prompt will ask value for.
	Variable string `json:"variable,omitempty"`
	// Prompt is the prompt text presented to the user when asking for value.
	//
	// On stdin the format will be: "Enter a value for <Prompt>".
	Prompt string `json:"prompt,omitempty"`
	// RegEx is the regular expression to use to validate the input string.
	// If RegEx is not set no validation will be performed on input in addition
	// to an empty value being accepted as a value.
	RegExp string `json:"regexp,omitempty"`
}

// Command defines a command to execute, either pre or post template execution.
type Command struct {
	// Name is the Command name.
	Name string `json:"name,omitempty"`
	// Program path to executable.
	Program string `json:"program,omitempty"`
	// Program arguments.
	// Placeholders in arguments are expended.
	Arguments []string `json:"arguments,omitempty"`
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
	// Prompts is a list of UserVariable prompts to present to the user
	// on Template execution.
	Prompts []*Prompt

	// directory is the directory from which metadata was loaded from.
	directory string
}

// errNoMetadata is returned by LoadMetadataFromDir if a metadata file
// does not exist in specified directory.
var errNoMetadata = errors.New("no metadata found")

// LoadMetadataFromDir loads metadata from dir and returns it or an error.
func LoadMetadataFromDir(dir string) (metadata *Metadata, err error) {
	var buf []byte
	if buf, err = ioutil.ReadFile(filepath.Join(dir, MetafileName)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errNoMetadata
		}
		return nil, fmt.Errorf("stat for metafile: %w", err)
	}
	metadata = &Metadata{directory: dir}
	if err = json.Unmarshal(buf, metadata); err != nil {
		return nil, fmt.Errorf("parse metafile: %w", err)
	}
	return
}

// Validate validates that metadata is properly formatted.
// It checks that multis point to valid Templates in the repo.
// It checks for duplicate template definitions.
// It checks for
func (self Metadata) Validate(repo Repository) error {
	return nil
}

// Metamap maps a template path to metadata loaded from its directory.
type Metamap map[string]*Metadata

// LoadMetamap loads metadata from root directory recursively and
// returns it or returns a descriptive error if one occured.
//
// The resulting Metamap will contain a key for each subdirectory
// recursively found in the repository. Only keys of paths to directories
// containing a Metafile, i.e. Template directories will have a valid
// *Metadata value. All other keys will have a nil value.
//
// The format of keys is a path relative to repo root i.e. 'apps/cli'.
// The key for the root is an empty string.
func LoadMetamap(root string) (metamap Metamap, err error) {
	var metadata *Metadata
	metamap = make(Metamap)
	err = filepath.Walk(root, func(path string, info fs.FileInfo, err error) error {
		if !info.IsDir() {
			return nil
		}
		// Add keys for Templates.
		metadata = nil
		if metadata, err = LoadMetadataFromDir(path); err != nil {
			if !errors.Is(err, errNoMetadata) {
				return fmt.Errorf("load metamap: %w", err)
			}
		}
		var s = strings.TrimPrefix(strings.TrimPrefix(path, root), string(os.PathSeparator))
		metamap[s] = metadata
		// Add keys for Multis.
		var base = filepath.Dir(s)
		if base == "." {
			base = ""
		}
		for _, multi := range metadata.Multis {
			s = filepath.Join(base, multi.Name)
			metamap[s] = metadata
		}
		return nil
	})
	return
}

// Metadata returns metadata for a path. If the path is invalid or no metadata
// for path exists an error is returned.
func (self Metamap) Metadata(path string) (*Metadata, error) {
	if strings.HasPrefix(path, string(os.PathSeparator)) {
		return nil, fmt.Errorf("metadata: invalid path: '%s'", path)
	}
	var meta, exists = self[path]
	if !exists {
		return nil, fmt.Errorf("no metadata for path '%s'", path)
	}
	return meta, nil
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
