// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package main implements boil's main executable.
package main

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
	"github.com/vedranvuk/boil/cmd/boil/internal/commands/edit"
	"github.com/vedranvuk/boil/cmd/boil/internal/commands/exec"
	"github.com/vedranvuk/boil/cmd/boil/internal/commands/info"
	"github.com/vedranvuk/boil/cmd/boil/internal/commands/list"
	"github.com/vedranvuk/boil/cmd/boil/internal/commands/newt"
	"github.com/vedranvuk/boil/cmd/boil/internal/commands/snap"
	"github.com/vedranvuk/cmdline"
)

// version is the boil version.
const version = "0.0.0-alpha"

var (
	err           error
	programConfig *boil.Config    // boil configuration
	cmdlineConfig *cmdline.Config // command line configuration
)

func main() {

	// Configuration defaults, later updated from file by the executed command.
	if programConfig, err = boil.DefaultConfig(); err != nil {
		fmt.Fprintf(os.Stderr, "init config: %s\n", err.Error())
		os.Exit(1)
	}
	// Command line configuration.
	cmdlineConfig = &cmdline.Config{
		Arguments:    os.Args[1:],
		NoAssignment: true,
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
				LongName: "version",
				Help:     "Show program version and exit.",
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
				MappedValue: &programConfig.Overrides.RepositoryPath,
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
			if err = programConfig.LoadOrCreate(); err != nil {
				return fmt.Errorf("configuration: %w", err)
			}
			if c.IsParsed("verbose") {
				fmt.Printf("Using configuration file: %s\n", programConfig.Runtime.LoadedConfigFile)
				programConfig.Print()
			}
			return nil
		},
		Commands: cmdline.Commands{
			{
				Name: "help",
				Help: "Show help, optionally for a specific topic.",
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
				Handler: handleHelp,
			},
			{
				Name: "list",
				Help: "List templates, optionally starting from specific subdirectory.",
				Options: cmdline.Options{
					&cmdline.Variadic{
						Name: "prefix",
						Help: "Start listing from this prefix.",
					},
				},
				Handler: func(c cmdline.Context) error {
					return list.Run(&list.Config{
						Prefix:        c.RawValues("prefix").First(),
						Configuration: programConfig,
					})
				},
			},
			{
				Name: "new",
				Help: "Create a new blank template and edit it.",
				Options: cmdline.Options{
					&cmdline.Boolean{
						LongName:  "overwrite",
						ShortName: "w",
						Help:      "Force overwrite if template exists",
					},
					&cmdline.Indexed{
						Name: "template-path",
						Help: "Name of the template to create.",
					},
				},
				Handler: func(c cmdline.Context) error {
					return newt.Run(&newt.Config{
						Overwrite:     c.IsParsed("overwrite"),
						TemplatePath:  c.RawValues("template-path").First(),
						Configuration: programConfig,
					})
				},
			},
			{
				Name: "snap",
				Help: "Create a new template from a source directory or file.",
				Options: cmdline.Options{
					&cmdline.Indexed{
						Name: "template-path",
						Help: "Path of the Template to be created.",
					},
					&cmdline.Boolean{
						LongName:  "wizard",
						ShortName: "z",
						Help:      "Define the template uzing a wizard.",
					},
					&cmdline.Boolean{
						LongName:  "overwrite",
						ShortName: "w",
						Help:      "Overwrite Template if it already exists without prompting.",
					},
					&cmdline.Variadic{
						Name: "source-path",
						Help: "Source directory or file path.",
					},
				},
				Handler: func(c cmdline.Context) error {
					return snap.Run(&snap.Config{
						TemplatePath:  c.RawValues("template-path").First(),
						Wizard:        c.IsParsed("wizard"),
						Overwrite:     c.IsParsed("overwrite"),
						SourcePath:    c.RawValues("source-path").First(),
						Configuration: programConfig,
					})
				},
			},
			{
				Name: "info",
				Help: "Show information about a template",
				Options: cmdline.Options{
					&cmdline.Indexed{
						Name: "template-path",
						Help: "Path of the template to show info for.",
					},
				},
				Handler: func(c cmdline.Context) error {
					return info.Run(&info.Config{
						TemplatePath:  c.RawValues("template-path").First(),
						Configuration: programConfig,
					})
				},
			},
			{
				Name: "edit",
				Help: "Edit a template using the default editor.",
				Options: cmdline.Options{
					&cmdline.Indexed{
						Name: "template-path",
						Help: "Path of the template to be edited.",
					},
				},
				Handler: func(c cmdline.Context) error {
					return edit.Run(&edit.Config{
						TemplatePath: c.RawValues("template-path").First(),
						EditAction:   "edit",
						Config:       programConfig,
					})
				},
				SubCommands: cmdline.Commands{
					{
						Name:    "all",
						Help:    "Edit all metafile values.",
						Handler: handleEditSubCommand,
					},
					{
						Name:    "info",
						Help:    "Edit basic metafile info.",
						Handler: handleEditSubCommand,
					},
					{
						Name: "file",
						Help: "Edit a template file. (Omit action to open file with editor)",
						Options: cmdline.Options{
							&cmdline.Indexed{
								Name: "file-path",
								Help: "Path of file to edit relative to template directory.",
							},
						},
						Handler: handleEditSubCommand,
						SubCommands: cmdline.Commands{
							{
								Name:    "touch",
								Help:    "Create a new file at specified path if it does not exist.",
								Handler: handleEditSubCommand,
								Options: cmdline.Options{
									&cmdline.Boolean{
										LongName:  "edit",
										ShortName: "e",
										Help:      "Open file with editor afterwards",
									},
								},
							},
							{
								Name:    "delete",
								Help:    "Delete a file at specified path.",
								Handler: handleEditSubCommand,
							},
						},
					},
					{
						Name: "directory",
						Help: "Edit a directory. (Omit action to open directory with editor)",
						Options: cmdline.Options{
							&cmdline.Indexed{
								Name: "directory-path",
								Help: "Path of the directory to edit relative to template directory.",
							},
						},
						Handler: handleEditSubCommand,
						SubCommands: cmdline.Commands{
							{
								Name:    "add",
								Help:    "Add a new directory at specified path.",
								Handler: handleEditSubCommand,
							},
							{
								Name:    "remove",
								Help:    "Remove a directory at specified path.",
								Handler: handleEditSubCommand,
								Options: cmdline.Options{
									&cmdline.Boolean{
										LongName:  "force",
										ShortName: "f",
										Help:      "Force removal of non-empty directories.",
									},
								},
							},
						},
					},
					{
						Name:    "prompts",
						Help:    "Edit prompts.",
						Handler: handleEditSubCommand,
					},
					{
						Name:    "preparse",
						Help:    "Edit pre-parse actions.",
						Handler: handleEditSubCommand,
					},
					{
						Name:    "preexec",
						Help:    "Edit pre-execute actions.",
						Handler: handleEditSubCommand,
					},
					{
						Name:    "postexec",
						Help:    "Edit post-execute actions.",
						Handler: handleEditSubCommand,
					},
					{
						Name:    "groups",
						Help:    "Edit template groups.",
						Handler: handleEditSubCommand,
					},
				},
			},
			{
				Name: "exec",
				Help: "Execute a template to a target directory.",
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
						Help:      "Print commands but do not execute them.",
					},
					&cmdline.Boolean{
						LongName:  "no-prompts",
						ShortName: "n",
						Help:      "Don't present input prompts for missing variables.",
					},
					&cmdline.Optional{
						LongName:  "output-dir",
						ShortName: "o",
						Help:      "Specify output directory (default: current directory).",
					},
					&cmdline.Repeated{
						LongName:  "var",
						ShortName: "r",
						Help:      "Define a variable.",
					},
				},
				Handler: func(c cmdline.Context) error {
					// Create a map of UserVariables.
					var vars = make(boil.Variables)
					for _, v := range c.RawValues("var") {
						var a = strings.Split(v, "=")
						if len(a) != 2 {
							return errors.New("variable must be in 'key=value' format")
						}
						vars[a[0]] = a[1]
					}
					// Execute Exec Command.
					return exec.Run(&exec.Config{
						TemplatePath:  c.RawValues("template-path").First(),
						OutputDir:     c.RawValues("output-dir").First(),
						Overwrite:     c.IsParsed("overwrite"),
						NoExecute:     c.IsParsed("no-execute"),
						NoPrompts:     c.IsParsed("no-prompts"),
						Vars:          vars,
						Configuration: programConfig,
					})
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
	if !cmdlineConfig.Commands.AnyExecuted() {
		cmdlineConfig.PrintUsage()
		os.Exit(0)
	}
}

// handleEditSubCommand handles the edit command and all of its subcommands.
func handleEditSubCommand(c cmdline.Context) error {
	return edit.Run(&edit.Config{
		TemplatePath:           c.GetParentCommand().Options.RawValues("template-path").First(),
		EditAction:             c.GetCommand().Name,
		ForceRemoveNonEmptyDir: c.IsParsed("force") && c.GetCommand().Name == "remove",
		EditAfterTouch:         c.IsParsed("edit") && c.GetCommand().Name == "touch",
		Config:                 programConfig,
	})
}
