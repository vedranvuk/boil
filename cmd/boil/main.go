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

const version = "0.0.0"

var (
	programConfig *boil.Config
	cmdlineConfig *cmdline.Config
	err           error
)

func main() {
	if programConfig, err = boil.DefaultConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "init config: %s\n", err.Error())
		os.Exit(1)
	}

	cmdlineConfig = &cmdline.Config{
		Arguments: os.Args[1:],
		GlobalsHandler: func(c cmdline.Context) (err error) {
			if c.IsParsed("help") {
				return handleHelp(c)
			}
			return nil
		},
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
				Name:    "exec",
				Help:    "Execute a template to a target directory.",
				Handler: handleExec,
				Options: cmdline.Options{
					&cmdline.Indexed{
						Name: "template-path",
						Help: "Path to a project template (local or remote).",
					},
					&cmdline.Optional{
						LongName:  "target-dir",
						ShortName: "t",
						Help:      "Target directory.",
					},
					&cmdline.Optional{
						LongName:  "project-name",
						ShortName: "n",
						Help:      "Specify project name.",
					},
					&cmdline.Boolean{
						LongName:  "no-create-dir",
						ShortName: "d",
						Help:      "Do not create a directory for te project, write directly to target-dir.",
					},
					&cmdline.Boolean{
						LongName:  "no-execute",
						ShortName: "x",
						Help:      "Do not execute commands.",
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
					&cmdline.Boolean{
						LongName:  "overwrite",
						ShortName: "w",
						Help:      "Overwrite any existing files in target directory.",
					},
				},
			},
			{
				Name:    "snap",
				Help:    "Create a new template from a directory snapshot.",
				Handler: handleSnap,
				Options: cmdline.Options{
					&cmdline.Boolean{
						LongName:  "confirm-files",
						ShortName: "c",
						Help:      "Prompt for each file if it should be included in the template.",
					},
					&cmdline.Boolean{
						LongName:  "force",
						ShortName: "f",
						Help:      "Force overwriting template if it already exists.",
					},
				},
			},
			{
				Name:    "list",
				Help:    "List templates, optionally in a specific subdirectory.",
				Handler: handleList,
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
				Name:    "info",
				Help:    "Show information about a template",
				Handler: handleInfo,
			},
		},
	}

	if err = cmdline.Parse(cmdlineConfig); err != nil {
		if errors.Is(err, cmdline.ErrNoArgs) {
			os.Exit(0)
		}
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func handleExec(c cmdline.Context) error {

	var vars = make(map[string]string)
	for _, v := range c.RawValues("var") {
		var a = strings.Split(v, "=")
		if len(a) != 2 {
			return errors.New("var must be in format key=value")
		}
		vars[a[0]] = a[1]
	}

	return exec.Run(&exec.Config{
		Config:        programConfig,
		TemplatePath:  c.RawValues("template-path").First(),
		ModulePath:    c.RawValues("module-path").First(),
		TargetDir:     c.RawValues("target-dir").First(),
		ProjectName:   c.RawValues("project-name").First(),
		NoExecute:     c.IsParsed("no-execute"),
		NoCreateDir:   c.IsParsed("no-create-dir"),
		Overwrite:     c.IsParsed("overwrite"),
		UserVariables: vars,
	})
}

func handleSnap(c cmdline.Context) error {
	return snap.Run(&snap.Config{
		ConfirmFiles: c.IsParsed("confirm-files"),
		Force:        c.IsParsed("force"),
	})
}

func handleList(c cmdline.Context) error {
	return list.Run(&list.Config{})
}

func handleInfo(c cmdline.Context) error {
	return list.Run(&list.Config{})
}
