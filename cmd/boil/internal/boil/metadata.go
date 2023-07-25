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

// MetafileName is the name of a file that defines a Boil template.
const MetafileName = "boil.json"

// Metafile is the Boil Template metadata. A directory with a valid Metafile
// defines a Template.
//
// If a Template contains other Templates in some of its subdirectories it can
// define one or more Multi Template definitions with various combinations of
// those child templates to be executed as part of the parent Template.
type Metafile struct {
	// Name is the Template name.
	// It is the last element of the template path when addressing it.
	// For example 'apps/<name>'
	Name string `json:"name,omitempty"`

	// Description is the Template description.
	// It is presented to the user when asking for template information.
	Description string `json:"description,omitempty"`

	// Author is the template author details.
	// This information is optional and is generated from the values set in
	// Configuration when generating a Template.
	Author *Author `json:"author,omitempty"`

	// Version is the template version, set manually used to help keep track of
	// Template changes. SemVer will be understood if this becomes important to
	// machine; currently just a meta field possibly useful to user.
	// By default the version is set at '1.0.0' when generating a Template.
	Version string `json:"version,omitempty"`

	// URL is an optional template url. Like Version, it has no meaning to the
	// machine but is just an additional meta field. It is empty by default.
	URL string `json:"url,omitempty"`

	// Files is a list of paths to files inside the Template directory that
	// will get executed and written to the output target directory retaining
	// its path relative to the Template directory.
	//
	// Paths must be relative to the Template directory and may not be rooted.
	// Files must point to existing files inside the Template directory.
	//
	// Directories for files will be created as needed, regardless of wether
	// they are defined in Directories.
	//
	// Paths of files defined in Files may contain placeholder values which will
	// get expended to actual values during Template execution.
	// A placeholder is defined with a "$" prefix, immediately followed by the
	// name of a Variable.
	Files []string `json:"files,omitempty"`

	// Directories is a list of directories to create in the target directory.
	// Placeholders are supported like with Files. Directories defined in this
	// list will be created regardless of wether they contain any of the
	// files defined by Files or if they exist phisically in the Template 
	// directory. They will be created in the template however when creating a
	// Template with the "snap" command.
	Directories []string `json:"directories,omitempty"`

	// Groups is a slice of Template Group definitions that may be executed
	// with the Template the metafile describes, as part of that Template.
	//
	// If the Template that this metafile describes contains other Templates
	// in any of its subdirectories, at any depths, one or more of those child
	// Templates may be combined into a named Group and addressed from it by a
	// path relative to this template.
	//
	// A Group can be executed along with the Template that defines it as part
	// of that template. This allows for defining segmented and multilayered
	// permutations of templates organized in a parent-child manner.
	//
	// A Group template is addressed by its path in a manner that the last
	// element of the path that matches Metafile.Name is instead replaced with
	// the name of the Group. So for instance, if a template 'apps/versatileapp'
	// defines groups 'base', and 'complete', to execute the 'base' Group the
	// path would be 'apps/base'.
	Groups []*Group `json:"groups,omitempty"`

	// Actions are groups of definitions of external actions to perform at
	// various stages of Template execution. In each Action group
	// (PreParse, PreExecute,...) the name of the Action must be unique and not
	// empty.
	Actions struct {
		// PreParse is a slice of actions to perform before any input variables
		// were parsed from any of sources defined on command line, in the
		// order they are defined. This is useful for a template setup like
		// temporary file generation, data input to variables, etc.
		//
		// No placeholders are available to expand in PerParse action
		// definition and any placeholder values found in the Action definition
		// will be unchanged and passed as defined, without raising an error.
		PreParse []*Action `json:"preParse,omitempty"`
		// PreExecute is a slice of actions to perform before the template is
		// executed in the order they are defined. It is called after the
		// variables were defined by parsing command line input, files given as
		// variable data on command line and all other input methods and are
		// available as expandable placeholders in action definition.
		//
		// This useful to execute some Template setup commands that depend on
		// Template variables.
		PreExecute []*Action `json:"preExecute,omitempty"`

		// PostExecute is a slice of actions to perform after the template was
		// executed, in order they are defined. This is useful for performing
		// cleanup operations. Variables will be available for expansion in the
		// action definition via placeholders.
		PostExecute []*Action `json:"postExecute,omitempty"`
	} `json:"actions,omitempty"`

	// Prompts is a list of prompts to present to the user before Template
	// execution via stdin to input values for variables the prompts define.
	//
	// Along with manually defining variables with the --var flag, a Template
	// can prompt the user for specific variables that the Template file needs.
	//
	// Prompts can each define a regular expression to use for input validation.
	// A failed validation will then re-prompt the user for value.
	Prompts []*Prompt

	// directory is the directory from which metadata was loaded from.
	directory string
}

// Author defines an author of a Template or a Repository.
type Author struct {
	// Name is the author name in an arbitrary format.
	Name string `json:"name,omitempty"`
	// Email is the author Email address.
	Email string `json:"email,omitempty"`
	// Homepage is the author's homepage URL.
	Homepage string `json:"homepage,omitempty"`
}

// Group defines a group of templates.
// See Metafile.Groups for details on Group usage.
type Group struct {
	// Name is the name of the Template Group.
	Name string `json:"name,omitempty"`
	// Description is the Group description text.
	Description string `json:"description,omitempty"`
	// Templates is a slice of Template names contained in this Group.
	Templates []string `json:"templates,omitempty"`
}

// Action defines some external action to execute via command line.
// See Metafile.Actions for details on Action usage.
type Action struct {
	// Description is the description text of the Action. It's an optional text
	// that should describe the action purpose.
	Description string `json:"description,omitempty"`
	// Program is the path to executable to run.
	Program string `json:"program,omitempty"`
	// Arguments are the arguments to pass to the executable.
	Arguments []string `json:"arguments,omitempty"`
	// WorkDir is the working directory to run the Program from.
	WorkDir string `json:"workDir,omitempty"`
	// Environment is the additional values to set in the Program environment.
	Environment map[string]string `json:"environment,omitempty"`
	// NoFail, if true will not break the execution of the process that ran
	// the Action, but it will generate a warning in the output.
	NoFail bool
}

// Prompt defines a prompt to the user for input of variable values.
// See Metafile.Prompts for details on Prompt usage.
type Prompt struct {
	// Variable is the name of the Variable this prompt will ask value for.
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

// errNoMetadata is returned by LoadMetadataFromDir if a metadata file
// does not exist in specified directory.
var errNoMetadata = errors.New("no metadata found")

// LoadMetadataFromDir loads metadata from dir and returns it or an error.
func LoadMetadataFromDir(dir string) (metadata *Metafile, err error) {
	var buf []byte
	if buf, err = ioutil.ReadFile(filepath.Join(dir, MetafileName)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errNoMetadata
		}
		return nil, fmt.Errorf("stat metafile: %w", err)
	}
	metadata = &Metafile{directory: dir}
	if err = json.Unmarshal(buf, metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metafile: %w", err)
	}
	return
}

// Validate validates that metadata is properly formatted.
// It checks that multis point to valid Templates in the repo.
// It checks for duplicate template definitions.
// It checks for
func (self Metafile) Validate(repo Repository) error {
	return nil
}

// Metamap maps a Template path to its Metadata.
type Metamap map[string]*Metafile

// MetamapFromDir loads metadata from root directory recursively recursively
// walking all child subdirectories and returns it or returns a descriptive
// error if one occurs.
//
// The resulting Metamap will contain a *Metadata for each subdirectory at any
// level in the root directory that contains a Metafile, i.e. defines a
// Template, under a key that will be a path relative to specified root.
//
// If root contains metadata i.e. is a Template itself an entry for it will
// be set under an empty key - not the standard current directory dot ".".
func MetamapFromDir(root string) (metamap Metamap, err error) {
	var metadata *Metafile
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
		if metadata != nil {
			var base = filepath.Dir(s)
			if base == "." {
				base = ""
			}
			if metadata.Groups != nil {
				for _, multi := range metadata.Groups {
					s = filepath.Join(base, multi.Name)
					metamap[s] = metadata
				}
			}
		}
		return nil
	})
	return
}

// Metadata returns metadata for a path. If the path is invalid or no metadata
// for path exists an error is returned.
func (self Metamap) Metadata(path string) (*Metafile, error) {
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
