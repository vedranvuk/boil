// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package main

import (
	"fmt"
	"os"
	"text/tabwriter"

	"github.com/vedranvuk/cmdline"
)

// handleHelp is the help command handler.
func handleHelp(c cmdline.Context) error {

	// List topics.
	if c.IsParsed("list-topics") {
		var wr = tabwriter.NewWriter(os.Stdout, 2, 2, 2, 32, 0)
		fmt.Fprintf(wr, "Available help topics are:\n\n")
		for _, topic := range helpTopics {
			fmt.Fprintf(wr, "\t%s\t%s\n", topic.Topic, topic.Description)
		}
		wr.Flush()
		return nil
	}

	// Show specific topic.
	if c.IsParsed("topic") {
		var topic = c.RawValues("topic").First()
		if !helpTopics.Exists(topic) {
			fmt.Printf("no help for '%s'\n", topic)
			os.Exit(1)
		}
		helpTopics.Print(topic)
		return nil
	}

	// Overview
	cmdline.PrintConfig(os.Stdout, cmdlineConfig)
	return nil
}

// HelpTopic defines a help topic.
type HelpTopic struct {
	// Topic is the keyword by which this help is referenced.
	Topic string
	// Description is a short overview text.
	Description string
	// Print prints the actual help text.
	Print func()
}

// HelpTopics is a slice of HelpTopic with few utility methods.
type HelpTopics []HelpTopic

// Exists returns true if a self contains a topic on keyword.
func (self HelpTopics) Exists(keyword string) bool {
	for _, t := range self {
		if t.Topic == keyword {
			return true
		}
	}
	return false
}

// Print prints the topic, if found.
func (self HelpTopics) Print(topic string) {
	for _, t := range self {
		if t.Topic == topic {
			t.Print()
			break
		}
	}
}

// helpTopics are the available help topic definitions.
var helpTopics = HelpTopics{
	{
		Topic:       "help",
		Description: "Help system usage.",
		Print:       printHelp,
	},
	{
		Topic:       "overview",
		Description: "Short overview on boil usage.",
		Print:       printOverview,
	},
	{
		Topic:       "repo",
		Description: "About repositories.",
		Print:       printRepository,
	},
	{
		Topic:       "metafile",
		Description: "Boil metafile reference.",
		Print:       printMetafile,
	},
	{
		Topic:       "bast",
		Description: "Bast reference.",
		Print:       printBast,
	},
	{
		Topic:       "globals",
		Description: "About global flags.",
		Print:       printGlobals,
	},
	{
		Topic:       "new",
		Description: "'new' command usage.",
		Print:       printNew,
	},
	{
		Topic:       "snap",
		Description: "'snap' command usage.",
		Print:       printSnap,
	},
	{
		Topic:       "list",
		Description: "'list' command usage.",
		Print:       printList,
	},
	{
		Topic:       "info",
		Description: "'info' command usage.",
		Print:       printInfo,
	},
	{
		Topic:       "edit",
		Description: "'edit' command usage.",
		Print:       printEdit,
	},
	{
		Topic:       "exec",
		Description: "'exec' command usage.",
		Print:       printExec,
	},
}

func printHelp() {
	fmt.Print(helpText)
}

func printOverview() {
	fmt.Print(overviewText)
}

func printRepository() {
	fmt.Print(repositoryText)
}

func printGlobals() {
	cmdline.PrintOptions(os.Stdout, cmdlineConfig, cmdlineConfig.Globals, 0)
	fmt.Print(globalsText)
}

func printNew() {
	cmdline.PrintCommand(os.Stdout, cmdlineConfig, cmdlineConfig.Commands.Find("new"), 0)
	fmt.Print(newText)
}

func printSnap() {
	cmdline.PrintCommand(os.Stdout, cmdlineConfig, cmdlineConfig.Commands.Find("snap"), 0)
	fmt.Print(snapText)
}

func printList() {
	cmdline.PrintCommand(os.Stdout, cmdlineConfig, cmdlineConfig.Commands.Find("info"), 0)
	fmt.Print(listText)
}

func printInfo() {
	cmdline.PrintCommand(os.Stdout, cmdlineConfig, cmdlineConfig.Commands.Find("info"), 0)
	fmt.Print(infoText)
}
func printEdit() {
	cmdline.PrintCommand(os.Stdout, cmdlineConfig, cmdlineConfig.Commands.Find("edit"), 0)
	fmt.Print(editText)
}

func printExec() {
	cmdline.PrintCommand(os.Stdout, cmdlineConfig, cmdlineConfig.Commands.Find("exec"), 0)
	fmt.Print(execText)
}

func printBast() {
	fmt.Print(bastText)
}

func printMetafile() {
	fmt.Print(metafileText)
}

const helpText = `Help command

The help command provides help on some topic or extended help about a command.
To get help about a specific command or topic type 'boil help <command|topic>'.

To list topics type 'boil help -l'
`

const overviewText = `Short primer on using boil


Boil

Boil is a tool for Go that takes a snapshot of a directory structure and 
packages it into a collection of patrametrized files and directories called
Templates which can then be used to create project boilerplates or smaller 
fragments of a project. The standard text/template package is used to enable 
parametrization of input files.


Template

A Template is a collection of parametrized files and directories packaged into 
a directory structure which is reflected in the output directory when executed.

File names of template files and directories can be parametrized using a simple
text substitution and content of template files is parametrized using
'text/template' package.

New templates are stored in a repository and boil maintains a default repository
in its configuration directory. A Template directory is identified by containing
a 'boil.json' Metafile which defines the Template. For more info on metafiles 
type 'boil help metafile'.

Data available to a template file can come from a variety of sources such as
standard input, command line arguments, AST of some input Go file or package 
(TODO: or various input files like json, yaml, xml, plain text files), etc.

For more info on data available to a template file see:

  'boil help exec' for info on how to access data from a template file.
  'boil help bast' for bast go parser reference.


Repository

Templates are created in Repositories which reside on disk in some directory.
Organization of templates in a repository is completely up to the user except
that template directories that contain one or more templates in any of their 
subdirectories can define groups which can execute one or more of those child 
templates as part of the parent template.

Templates are addressed by path inside the Repository, i.e.: 'apps/webapp' or 
by using an absolute path to a Template directory, i.e.: '/home/templates/app'
in which case repository is ignore and the template loaded directly from the
specified directory.


Template paths

A template path is a simple relative path that addresses a template directory
inside a repository. This format is used by all except the exec command.

A template path relative to repository:

  go/apps/cmdapp

The exec command supports an extended template path that may address a template
outside of the loaded repository by using an absolute path. Additonally, the 
exec command supports a URL fragment like suffix that names a group to execute 
defined in a template.

An absolute path to a template:

  /home/user/templates/apptemplate

An absolute path to a template that addresses a group defined in the template:

  /home/user/templates/apptemplate#all

For more info on template paths and groups see 'boil help repository'.
`

const repositoryText = `
Repository

A Repository is any directory that contains Templates, possibly organized in a
manner customized by the user. 

The simplest example is a repository containing a single template named 'foo' 
defined by its boil.json file and containing a single file in a subdirectory: 

  /repository
    /foo
      /cmd
        main.go
      boil.json

The 'foo' template would in this case be addressed with 'foo'.

User may categorize templates when defining them by organizing them into 
subdirectories by prepending the template name with some path prefix which will 
be reflected in the repository. For example a template 'go/foo' would be stored
in the repository as:

  /repository
    /go
      /foo
        /cmd
          main.go
        boil.json

A template may contain one or more other templates in its subdirectories and if
it does it may contain one or more group definitions which specify which of the
child templates will get executed as part of the parent template. Take for 
instance the following repository:

  /repository
    /foo
      /cmd
        main.go
      /config
        /cmd
          config.go
        boil.json
      /webui
        /cmd
          webui.go
        boil.json
      boil.json
	
In this example repository contains a single template named 'foo' which has two 
child templates named 'config' and 'webui'. The parent 'foo' template can then 
define one or more groups, each of which can reference various child template
combinations whose files will be executed along with the parent template files
to the same output directory specified to the exec command.

A group is referenced by appending '#' to a template path immediately followed 
by the name of the group, e.g. 'foo#all'.

Say the 'foo' template metafile '/foo/boil.json' defines two groups: a 'config' 
group which references only the config 'child' template and an 'all' group which
references the "config' and 'webui' templates.

Executing 'foo#config' would along with the 'foo' template files also execute
the files of 'config' template in the same output directory and executing
'foo#all' would execute files contained by the 'webui' template as well.

Templates referenced by the group are executed after the parent template files
and in the order as they are defined in the metafile.
`

const metafileText = `Metafile

TODO: Metafile help.
`

const bastText = `Bast

(B)astard (AST) defines a simple object model from standard Go AST which allows
easier access to an input go file syntax tree and is designed to be used from 
within a template file being executed using 'text/template'.

Currently, it parses only top level interface and struct declarations from each 
input file.

It is accessible from {{.Bast}} pipeline from inside a template file or via 
template functions.

TODO: BAST object reference.
TODO: BAST function reference.
`

const globalsText = `
About --no-repository

This option disables the use of repositories. All template paths will be 
considered as paths pointing to a root of a template, relative or absolute. 
`

const newText = `
Usage: boil new <template-path> [options]

The new command creates a new template.
`

const snapText = `
Usage: boil snap <template-path> [options]

The new command creates a new template from a source directory.
`

const listText = `
Usage: boil edit <template-path> [options] [subcommand [options]]

The edit commands edits a template files or template metadata.

Executing the edit command without a subcommand opens the template directory in
the editor defined in the configuration file. If one is not defined the template
directory is opened in the system default file explorer.

Edit subcommands open command prompt editors for metadata or parts of it.
`

const infoText = `
Usage: boil edit <template-path> [options] [subcommand [options]]

The edit commands edits a template files or template metadata.

Executing the edit command without a subcommand opens the template directory in
the editor defined in the configuration file. If one is not defined the template
directory is opened in the system default file explorer.

Edit subcommands open command prompt editors for metadata or parts of it.
`

const editText = `
Usage: boil edit <template-path> [options] [subcommand [options]]

The edit commands edits a template files or template metadata.

Executing the edit command without a subcommand opens the template directory in
the editor defined in the configuration file. If one is not defined the template
directory is opened in the system default file explorer.

Edit subcommands open command prompt editors for metadata or parts of it.
`

const execText = `
Usage: boil exec <template-path> [options]

The exec command executes a template by copying its files and directories to the
output directory retaining directory structure. It replaces variable 
placeholders in source template file names in the process and passes data 
defined on the command line or some other input to each template file as it is 
executed to its output location.

Exec command executes each entry in the order as defined for each set of actions
defined in the template metafile at following stages of exec command:

 PreParse: Before variable parsing or any template file enumeration. May be used
           for some external setup or similar.

 PreExec:  Just before template file executions, after all variables have been 
           loaded. Useful for some external input generation or similar.

 PostExec: After template file executions, useful for cleanup of anything 
           generated using earlier actions.

Any prompts defined in the template will be presented to the user to enter
values for variables they define via stdin dialogs unless '--no-prompt' is given
in command line arguments.

If a variable defined by a template prompt is defined on the command line using 
the 'var' option the prompt for it will not be presented to the user.

If '--no-prompts' is specified in command line arguments no prompts defined in 
the template will be presented to the user but if a variable defined by a prompt
is not oterwise given using the 'var' exec option the exec command will fail.

Variables defined using the 'var' option can (currently) also override values 
of exec option values, and take precedence over values given in options 
themselves. Var option can be specified multiple times. 
The mapping is as follows: 

  OutputDirectory  output-dir

The 'go-input' option takes a relative or absolute path to a go package or a 
file and can be specified multiple times. Each package will be a separate 
package in the resulting Bast and all files will be in a placeholder package
which can be referenced from a template by using an empty string as the package 
name in functions that require it.

The 'no-metadata' option disables use of template metadata so any functions
that are supported by the metafile will not function. This includes prompts, 
template groups and actions. All variables required by template files must be
declared up fron using the 'var' option or the execution will fail.

Variable placeholders still work and the source template is copied recursively
to the output directory.
`
