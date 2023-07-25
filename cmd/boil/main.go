package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
	"github.com/vedranvuk/boil/cmd/boil/internal/commands/exec"
	"github.com/vedranvuk/boil/cmd/boil/internal/commands/list"
	"github.com/vedranvuk/boil/cmd/boil/internal/commands/snap"
	"github.com/vedranvuk/cmdline"
)

// version is the boil version.
const version = "0.0.0-alpha"

var (
	programConfig *boil.Configuration // boil configuration
	cmdlineConfig *cmdline.Config     // command line configuration
	err           error               // reusable error
)

func main() {

	// Configuration defaults, later updated from file by the executed command.
	if programConfig, err = boil.DefaultConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "init config: %s\n", err.Error())
		os.Exit(1)
	}
	// Command line configuration.
	cmdlineConfig = &cmdline.Config{
		Arguments: os.Args[1:],
		Globals: cmdline.Options{
			&cmdline.Boolean{
				LongName:  "help",
				ShortName: "h",
				Help:      "Show help.",
			},
			&cmdline.Boolean{
				LongName:    "verbose",
				ShortName:   "v",
				Help:        "Enable verbose output.",
				MappedValue: &programConfig.Overrides.Verbose,
			},
			&cmdline.Boolean{
				LongName:    "prompt",
				ShortName:   "p",
				Help:        "Prompt for missing required arguments via stdin.",
				MappedValue: &programConfig.Overrides.Prompt,
			},
			&cmdline.Optional{
				LongName:    "config",
				ShortName:   "c",
				Help:        "Override filename of config file to use.",
				MappedValue: &programConfig.Overrides.ConfigFile,
			},
			&cmdline.Optional{
				LongName:    "repository",
				ShortName:   "r",
				Help:        "Override directory of repository to use.",
				MappedValue: &programConfig.Overrides.Repository,
			},
			&cmdline.Boolean{
				LongName: "version",
				Help:     "Show program version and exit.",
			},
		},
		GlobalExclusivityGroups: []cmdline.ExclusivityGroup{
			{
				"verbose",
				"version",
				"help",
			},
		},
		GlobalsHandler: func(c cmdline.Context) (err error) {
			if c.IsParsed("help") {
				return handleHelp(c)
			}
			if c.IsParsed("version") {
				fmt.Printf("boil v%s\n", version)
				os.Exit(0)
			}
			return nil
		},

		Commands: cmdline.Commands{
			{
				Name:    "help",
				Help:    "Show help, optionally for a specific topic.",
				Handler: handleHelp,
				Options: cmdline.Options{
					&cmdline.Boolean{
						LongName:  "list-topics",
						ShortName: "l",
						Help:      "List help topics.",
					},
					&cmdline.Variadic{
						Name: "topic",
						Help: "Help topic to display.",
					},
				},
			},
			{
				Name: "list",
				Help: "List templates, optionally starting from specific subdirectory.",
				Handler: func(c cmdline.Context) error {
					return list.Run(&list.Config{})
				},
				Options: cmdline.Options{
					&cmdline.Boolean{
						LongName:  "recursive",
						ShortName: "r",
						Help:      "List templates recursively.",
					},
					&cmdline.Variadic{
						Name: "path",
						Help: "Template subdirectory path to list.",
					},
				},
			},
			{
				Name: "snap",
				Help: "Create a new template from a source directory or file.",
				Handler: func(c cmdline.Context) error {
					return snap.Run(&snap.Config{
						ConfirmFiles: c.IsParsed("confirm-files"),
						Force:        c.IsParsed("force"),
					})
				},
				Options: cmdline.Options{
					&cmdline.Boolean{
						LongName:  "overwrite",
						ShortName: "w",
						Help:      "Overwrite Template if it already exists without prompting.",
					},
					&cmdline.Boolean{
						LongName:  "confirm-files",
						ShortName: "c",
						Help:      "Prompt for each file if it should be included in the template.",
					},
					&cmdline.Boolean{
						LongName:  "wizard",
						ShortName: "z",
						Help:      "Define additional Template properties using a wizard.",
					},
				},
			},
			{
				Name: "info",
				Help: "Show information about a template",
				Handler: func(c cmdline.Context) error {
					return list.Run(&list.Config{})
				},
			},
			{
				Name: "edit",
				Help: "Edit a template using the default editor.",
				Handler: func(c cmdline.Context) error {
					return nil
				},
				Options: cmdline.Options{
					&cmdline.Indexed{
						Name: "template-path",
						Help: "Path f the Template to be edited.",
					},
				},
			},
			{
				Name: "exec",
				Help: "Execute a template to a target directory.",
				Handler: func(c cmdline.Context) error {
					// Create a map of UserVariables.
					var vars = make(map[string]string)
					for _, v := range c.RawValues("var") {
						var a = strings.Split(v, "=")
						if len(a) != 2 {
							return errors.New("var must be in format key=value")
						}
						vars[a[0]] = a[1]
					}
					// Execute Exec Command.
					return exec.Run(&exec.Config{
						TemplatePath:  c.RawValues("template-path").First(),
						TargetDir:     c.RawValues("target-dir").First(),
						NoExecute:     c.IsParsed("no-execute"),
						Overwrite:     c.IsParsed("overwrite"),
						UserVariables: vars,
						Configuration: programConfig,
					})
				},
				Options: cmdline.Options{
					&cmdline.Indexed{
						Name: "template-path",
						Help: "Path of the Template to be executed.",
					},
					&cmdline.Boolean{
						LongName:  "overwrite",
						ShortName: "w",
						Help:      "Overwrite any existing output files without prompting.",
					},
					&cmdline.Boolean{
						LongName:  "no-execute",
						ShortName: "x",
						Help:      "Do not execute commands.",
					},
					&cmdline.Optional{
						LongName:  "target-dir",
						ShortName: "t",
						Help:      "Target directory.",
					},
					&cmdline.Optional{
						LongName:  "module-path",
						ShortName: "m",
						Help:      "Module path to use for generating go.mod files.",
					},
					&cmdline.Repeated{
						LongName:  "var",
						ShortName: "r",
						Help:      "Define a variale addressable from templates.",
					},
				},
			},
		},
	}
	// Parse command line.
	if err = cmdline.Parse(cmdlineConfig); err != nil {
		if errors.Is(err, cmdline.ErrNoArgs) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, fmt.Errorf("error: %w", err))
		os.Exit(1)
	}
}
