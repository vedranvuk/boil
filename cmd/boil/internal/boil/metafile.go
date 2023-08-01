package boil

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/tabwriter"
)

// MetafileName is the name of a file that defines a Boil template.
const MetafileName = "boil.json"

// NewMetafile returns a new metfile initialized to defaults from config.
func NewMetafile(config *Config) *Metafile {
	return &Metafile{
		Author: &Author{
			Name:     config.DefaultAuthor.Name,
			Email:    config.DefaultAuthor.Email,
			Homepage: config.DefaultAuthor.Homepage,
		},
		Version:     "1.0.0",
		URL:         "https://",
		Directories: []string{},
		Files:       []string{},
		Prompts:     Prompts{},
		Groups:      []*Group{},
	}
}

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
	Files []string `json:"files"`

	// Directories is a list of directories to create in the target directory.
	// Placeholders are supported like with Files. Directories defined in this
	// list will be created regardless of wether they contain any of the
	// files defined by Files or if they exist phisically in the Template
	// directory. They will be created in the template however when creating a
	// Template with the "snap" command.
	Directories []string `json:"directories"`

	// Prompts is a list of prompts to present to the user before Template
	// execution via stdin to input values for variables the prompts define.
	//
	// Along with manually defining variables with the --var flag, a Template
	// can prompt the user for specific variables that the Template file needs.
	//
	// Prompts can each define a regular expression to use for input validation.
	// A failed validation will then re-prompt the user for value.
	Prompts Prompts `json:"prompts,imitempty"`

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
		PreParse Actions `json:"preParse,omitempty"`
		// PreExecute is a slice of actions to perform before the template is
		// executed in the order they are defined. It is called after the
		// variables were defined by parsing command line input, files given as
		// variable data on command line and all other input methods and are
		// available as expandable placeholders in action definition.
		//
		// This useful to execute some Template setup commands that depend on
		// Template variables.
		PreExecute Actions `json:"preExecute,omitempty"`

		// PostExecute is a slice of actions to perform after the template was
		// executed, in order they are defined. This is useful for performing
		// cleanup operations. Variables will be available for expansion in the
		// action definition via placeholders.
		PostExecute Actions `json:"postExecute,omitempty"`
	} `json:"actions,omitempty"`

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

	// Path is where metafile resides, relative to the repository root.
	// it is equal to template Path minus the optional group name.
	//
	// Path is not stored with the template, it's runtime only.
	Path string `json:"-"`
}

func (self *Metafile) Print() {
	var wr = tabwriter.NewWriter(os.Stdout, 2, 2, 2, 32, 0)
	fmt.Fprintf(wr, "Name:\t%s\n", self.Name)
	fmt.Fprintf(wr, "Description:\t%s\n", self.Description)
	fmt.Fprintf(wr, "Author Name:\t%s\n", self.Author.Name)
	fmt.Fprintf(wr, "Author Email:\t%s\n", self.Author.Email)
	fmt.Fprintf(wr, "Author Homepage:\t%s\n", self.Author.Homepage)
	fmt.Fprintf(wr, "Version:\t%s\n", self.Version)
	fmt.Fprintf(wr, "URL:\t%s\n", self.URL)
	fmt.Fprintf(wr, "Directories:\t\n")
	for _, dir := range self.Directories {
		fmt.Fprintf(wr, "\t%s\n", dir)
	}
	fmt.Fprintf(wr, "Files:\t\n")
	for _, file := range self.Files {
		fmt.Fprintf(wr, "\t%s\n", file)
	}
	fmt.Fprintf(wr, "Prompts:\t\n")
	for _, prompt := range self.Prompts {
		fmt.Fprintf(wr, "Variable:\t%s\n", prompt.Variable)
		fmt.Fprintf(wr, "Description:\t%s\n", prompt.Description)
		fmt.Fprintf(wr, "RegExp:\t%s\n", prompt.RegExp)
	}
	fmt.Fprintf(wr, "PreParse Actions:\t\n")
	for _, action := range self.Actions.PreParse {
		fmt.Fprintf(wr, "Description:\t%s\n", action.Description)
		fmt.Fprintf(wr, "Program:\t%s\n", action.Program)
		fmt.Fprintf(wr, "Arguments:\t%v\n", action.Arguments)
		fmt.Fprintf(wr, "WorkDir:\t%s\n", action.WorkDir)
		fmt.Fprintf(wr, "NoFailt%t\n", action.NoFail)
	}
	fmt.Fprintf(wr, "PreExecute Actions:\t\n")
	for _, action := range self.Actions.PreExecute {
		fmt.Fprintf(wr, "Description:\t%s\n", action.Description)
		fmt.Fprintf(wr, "Program:\t%s\n", action.Program)
		fmt.Fprintf(wr, "Arguments:\t%v\n", action.Arguments)
		fmt.Fprintf(wr, "WorkDir:\t%s\n", action.WorkDir)
		fmt.Fprintf(wr, "NoFailt%t\n", action.NoFail)
	}
	fmt.Fprintf(wr, "PostExecute Actions:\t\n")
	for _, action := range self.Actions.PostExecute {
		fmt.Fprintf(wr, "Description:\t%s\n", action.Description)
		fmt.Fprintf(wr, "Program:\t%s\n", action.Program)
		fmt.Fprintf(wr, "Arguments:\t%v\n", action.Arguments)
		fmt.Fprintf(wr, "WorkDir:\t%s\n", action.WorkDir)
		fmt.Fprintf(wr, "NoFailt%t\n", action.NoFail)
	}
	fmt.Fprintf(wr, "Groups:\t\n")
	for _, group := range self.Groups {
		fmt.Fprintf(wr, "Name:\t%s\n", group.Name)
		fmt.Fprintf(wr, "Description:\t%s\n", group.Description)
		fmt.Fprintf(wr, "Templates:\t%v\n", group.Templates)
	}
	wr.Flush()
}

// errNoMetadata is returned by LoadMetadataFromDir if a metadata file
// does not exist in specified directory.
var errNoMetadata = errors.New("no metadata found")

// LoadMetafileFromDir loads metadata from dir and returns it or an error.
func LoadMetafileFromDir(dir string) (metadata *Metafile, err error) {
	var buf []byte
	if buf, err = os.ReadFile(filepath.Join(dir, MetafileName)); err != nil {
		if errors.Is(err, os.ErrNotExist) {
			return nil, errNoMetadata
		}
		return nil, fmt.Errorf("stat metafile: %w", err)
	}
	metadata = new(Metafile)
	if err = json.Unmarshal(buf, metadata); err != nil {
		return nil, fmt.Errorf("unmarshal metafile: %w", err)
	}
	return
}

// NewAuthor returns a new *Author.
func NewAuthor() *Author { return &Author{} }

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

// Prompt defines a prompt to the user for input of variable values.
// See Metafile.Prompts for details on Prompt usage.
type Prompt struct {
	// Variable is the name of the Variable this prompt will ask value for.
	Variable string `json:"variable,omitempty"`
	// Description is the prompt text presented to the user when asking for value.
	//
	// On stdin the format will be: "Enter a value for <Description>".
	Description string `json:"description,omitempty"`
	// RegEx is the regular expression to use to validate the input string.
	// If RegEx is not set no validation will be performed on input in addition
	// to an empty value being accepted as a value.
	RegExp string `json:"regexp,omitempty"`
	// Optional if true will not trigger an error if the variable was assigned 
	// an empty value.
	Optional bool `json:"optional,omitempty"`
}

// Prompts is a slice of *Prompt.
type Prompts []*Prompt

// FindByVariable returns a Prompt that defines variable or nil if not found.
func (self Prompts) FindByVariable(variable string) *Prompt {
	for _, prompt := range self {
		if prompt.Variable == variable {
			return prompt
		}
	}
	return nil
}

// Validate validates that metadata is properly formatted.
// It checks that multis point to valid Templates in the repo.
// It checks for duplicate template definitions.
func (self Metafile) Validate(repo Repository) error {
	// TODO Implement Metafile.Validate
	return nil
}

// ExecPreParseActions executes all PreParse Actions defined in the Metafile.
// It returns the error of the first Action that failed and stops execution.
// If no error occurs nil is returned.
func (self Metafile) ExecPreParseActions() error {
	return self.Actions.PreParse.ExecuteAll(nil)
}

// ExecPreExecuteActions executes all PreExecute Actions defined in the Metafile.
// It returns the error of the first Action that failed and stops execution.
// If no error occurs nil is returned.
func (self Metafile) ExecPreExecuteActions(variables Variables) error {
	return self.Actions.PreExecute.ExecuteAll(variables)
}

// ExecPostExecuteActions executes all PostExecute Actions defined in the Metafile.
// It returns the error of the first Action that failed and stops execution.
// If no error occurs nil is returned.
func (self Metafile) ExecPostExecuteActions(variables Variables) error {
	return self.Actions.PostExecute.ExecuteAll(variables)
}

// Metamap maps a Template path to its Metadata.
type Metamap map[string]*Metafile

// Metafile returns metadata for a path. If the path is invalid or no metadata
// for path exists an error is returned.
func (self Metamap) Metafile(path string) (*Metafile, error) {
	if strings.HasPrefix(path, string(os.PathSeparator)) {
		return nil, fmt.Errorf("metadata: invalid path: '%s'", path)
	}
	var meta, exists = self[path]
	if !exists {
		return nil, os.ErrNotExist
	}
	return meta, nil
}

// Print printes self to stdout.
func (self Metamap) Print() {
	var wr = tabwriter.NewWriter(os.Stdout, 2, 2, 2, 32, 0)
	var a []string
	for k := range self {
		a = append(a, k)
	}
	sort.Strings(a)
	fmt.Fprintf(wr, "[Template Name]\t[Path]\t[Description]\n")
	for _, v := range a {
		var s = "nil"
		if self[v] != nil {
			s = self[v].Name
		}
		fmt.Fprintf(wr, "%s\t%s\t%s\n", s, v, self[v].Description)
	}
	fmt.Fprintf(wr, "\n")
	wr.Flush()
}
