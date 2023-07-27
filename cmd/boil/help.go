package main

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/vedranvuk/cmdline"
)

func handleHelp(c cmdline.Context) error {

	// List topics.
	if c.IsParsed("list-topics") {
		if c.IsParsed("topic") {
			return errors.New("Options 'list-topics' and 'topic' are mutually exclusive")
		}
		var wr = tabwriter.NewWriter(os.Stdout, 2, 2, 2, 32, 0)
		fmt.Fprintf(wr, "Available help topics are:\n\n")
		for k, v := range helpTopics {
			var title, err = bufio.NewReader(strings.NewReader(v)).ReadString('\n')
			if err != nil {
				panic(fmt.Errorf("unable to scan help topic title: %w", err))
			}
			fmt.Fprintf(wr, "\t%s\t%s", k, title)
		}
		wr.Flush()
		return nil
	}

	// Show specific topic.
	if c.IsParsed("topic") {
		var topic = c.RawValues("topic").First()
		var text, exists = helpTopics[topic]
		if !exists {
			fmt.Printf("no help for '%s'\n", topic)
			os.Exit(1)
		}
		fmt.Print(text)
		return nil
	}

	// Overview
	fmt.Println()
	cmdline.PrintConfig(os.Stdout, cmdlineConfig)
	return nil
}

// helpTopics is a map of topic keywords to topic text.
//
// Formatting of the topic text must be in the form of a:
// * Descriptive title on first line that will also be used in topic listing.
// * A blank line.
// * The topic text itself, further formatted freely.
//
// If these rules are not met, listing help topics via 'help -l' will panic.
var helpTopics = map[string]string{
	"overview": overviewTopic,
	"help":     helpTopic,
	"exec":     execTopic,
	"stdvars":  stdvarsTopic,
}

const (
	overviewTopic = `Short primer on using boil

Boil

Boil is a tool for Go that takes a snapshot of a directory structure and 
packages it into a collection of patrametrized files and directories called
Templates which can then be used to create project boilerplates or smaller 
fragments of a project. The standard text/template package is used to enable 
parametrization of input files.

Data available to a template file can come from a variety of sources such as
standard input, command line arguments, AST of some input Go file or package or
various input files like json, yaml, xml, plain text files, etc.

Templates are created in Repositories which reside on disk in some directory
structured with a layout that Boil recognizes and maintains. Such Repositories
can be easily versioned with git and specified as overrides on the command line
when working with boil.

Templates are addressed by path inside the Repository, i.e.: 'apps/webapp' or 
by using an absolute path to a Template directory, i.e.: '/home/templates/app'.


Template

A Template is a collection of parametrized files and directories packaged in a 
directory structure which is reflected in the output directory when executed.

A Template directory is identified by containing a 'boil.json' Metafile which 
defines the Template.


Repository

A Repository is any directory that contains Templates, possibly organized in a
manner customized by the user. Take for example a simple Repository structure:

/repository
	/apps
   		/cliapp
			/cmd
				/app
					main.go
			boil.json
   		/webapp
			/cmd
				/app
					main.go
			boil.json
	/multis
		/segmented
			/docs
				manual.md
			/base
				/cmd
					/app
						main.go
				boil.json
			/config
				/internal
					config.go
				boil.json
			/webui
				/internal
					webui.go
				boil.json
			/api
				/internal
					api.go
				boil.json
			boil.json
			README.md
	
In this example 'repository' is the root of some Repository directory. It
contains two directories 'apps' and 'multis' which have no metadata and serve 
only to categorize Templates by some arbitrary hierarchy defined by user.

Inside the 'apps' directory are two subdirectories 'cliapp' and 'webapp' with 
their Metadata files, so those two directories each define a single Template.
All subdirectories and files of a Template that are listed in the Metafile
will be executed in the target directory retaining directory structure.

To execute for instance the 'webapp' template one would need to qualify it 
by path, e.g.: 'apps/webapp'

In the 'multis' directory there is a Multi Template named 'segmented' defined
by its Metafile. It may define a list of its files and subdirectories to be 
executed with the Multi, in this case directory 'docs', and the files in the 
'docs' and the 'README.md' as well as any number of combinations of Templates
defined in its subdirectories ('base', 'config', 'webui', 'api').

The name of Template in a Template path (last element) can then be substituted
for the name of a Multi defined in the Metafile of a Template, e.g.:

'multis/segmented/base_with_api' - for instance, files in 'segmented' and the 
'base' and 'api' templates all executed as a single Template...

'multis/segmented/complete' - ...or all files and subdirectories in 'segmented' 
along with all Templates defined in subdirectories of 'segmented'.

The Templates defined inside a Multi can also be addressed directly, e.g.
'multis/segmented/base' and executed into some target directory separately.

Executing a template by name that contains Multi definitions,  e.g.: 
'multis/segmented' will only execute files and directories defined in 
the 'segmented' Metafile but not any of the Templates defined in subdirectories.


Metafile

A Metafile defines metadata about the author, version and contact details and
the list of files and directories comprising the Template which are to be 
executed into target directory.

	// Metadata represents a Template Metafile.
	type Metadata struct {
		Name        string   // Template name.
		Description string   // Description.
		Author      *Author  // Template Author information.
		Version     string   // Template version.
		URL         string   // URL is the cannonical template URL.
		Files       []string // List of Template files.
		Directories []string // List of Template directories.
		Multis      []*Multi // Multi definition.
		Actions     struct {
			Pre  []*Command // Pre-execution commands.
			Post []*Command // Post execution commands.
		} // Custom actions to execute with Template.
	}

	// Multi defines a a Multi Template.
	type Multi struct {
		Name        string   // Name is the Multi.
		Description string   // Description.
		Templates   []string // Templates to execute as part of Multi.
	}

	// Author defines an author.
	type Author struct {
		Name 		string 	// Name is the author name in an arbitrary format.
		Email 		string 	// Email is the author Email address.
		Homepage 	string 	// Homepage is the author's homepage URL.
	}

	// Command defines a command to execute
	type Command struct {
		Name 		string 		// Name is the Command name.
		Program 	string 		// Program path to executable.
		Arguments 	[]string 	// Program arguments.
	}

`

	helpTopic = `Help command

The help command provides extended help about a command.
To get help about a specific command type 'boil help <command>'.
`

	execTopic = `Exec command
	
The exec command executes a template to a target directory replacing placeholder 
variables in the process and executing a template using values provided on 
command line or extracted from a go file.
`

	stdvarsTopic = `Standard variables

The standard variables available to all template files during template execution
including file name expansion are:

	OutputDirectory	 Absolute path of the output directory.
	TemplatePath     Template path as specified.
`
)
