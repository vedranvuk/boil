// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package edit implements boil's edit command.
package edit

import (
	"fmt"
	"path/filepath"
	"strings"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Config is the Edit command configuration.
type Config struct {
	// TemplatePath is the path of the template to edit.
	// It may not contain group names.
	TemplatePath string
	// EditAction specifies the edit sub action.
	EditAction string
	// EditTarget is the target file or dir for some edit actions.
	EditTarget string
	// ForceRemoveNonEmptyDir removal of non-empty directories.
	ForceRemoveNonEmptyDir bool
	// Open the file with editor after touching it.
	EditAfterTouch bool
	// Config is the loaded program configuration.
	Config *boil.Config
}

type state struct {
	config   *boil.Config
	repo     boil.Repository
	meta     *boil.Metafile
	vars     boil.Variables
	tmplPath string
	repoPath string
}

// Run executes the Edit command configured by config.
// If an error occurs it is returned and the operation may be considered failed.
func Run(config *Config) (err error) {

	fmt.Printf("%#v\n", config)

	var state = &state{
		config:   config.Config,
		vars:     make(boil.Variables),
		tmplPath: config.TemplatePath,
		repoPath: config.Config.GetRepositoryPath(),
	}

	state.tmplPath, _, _ = strings.Cut(config.TemplatePath, "#")

	if filepath.IsAbs(config.TemplatePath) {
		// If TemplatePath is an absolute path open the Template as the
		// Repository and adjust the template path to "current directory"
		// pointing to repository root.
		state.repoPath = state.tmplPath
		state.tmplPath = "."
		if config.Config.Overrides.Verbose {
			fmt.Println("Absolute Template path specified, repository opened at template root.")
		}
	}

	if state.repo, err = boil.OpenRepository(state.repoPath); err != nil {
		return fmt.Errorf("open repository: %w", err)
	}
	if state.meta, err = state.repo.OpenMeta(state.tmplPath); err != nil {
		return fmt.Errorf("template %s not found", config.TemplatePath)
	}

	state.vars[boil.VarTemplatePath.String()] = filepath.Join(state.repo.Location(), config.TemplatePath)

	var (
		tgtExists, entryExists bool
		absTarget string
	)
	switch config.EditAction {
	case "edit":
		state.vars[boil.VarEditTarget.String()] = filepath.Join(state.repo.Location(), config.TemplatePath)
		return config.Config.ExternalEditor.Execute(state.vars)
	case "all":
		err = boil.NewEditor(config.Config, state.meta).EditAll()
	case "info":
		err = boil.NewEditor(config.Config, state.meta).EditInfo()
	case "files":
		err = boil.NewEditor(config.Config, state.meta).EditFiles()
	case "dirs":
		err = boil.NewEditor(config.Config, state.meta).EditDirs()
	case "prompts":
		err = boil.NewEditor(config.Config, state.meta).EditPrompts()
	case "preparse":
		err = boil.NewEditor(config.Config, state.meta).EditPreParse()
	case "preexec":
		err = boil.NewEditor(config.Config, state.meta).EditPreExec()
	case "postexec":
		err = boil.NewEditor(config.Config, state.meta).EditPostExec()
	case "groups":
		err = boil.NewEditor(config.Config, state.meta).EditGroups()
	case "addFile":
		absTarget = filepath.Join(state.repo.Location(), state.tmplPath, config.EditTarget)
		if tgtExists, err = state.repo.Exists(absTarget); err != nil {
			return
		}
		for _, entry := range state.meta.Files {
			if strings.EqualFold(entry, config.EditTarget) {
				entryExists = true
				break
			}
		}
		if tgtExists && entryExists {
			fmt.Printf("file '%s' already exists\n", config.EditTarget)
			return nil
		}
		fmt.Println("addFile")
	case "remFile":
		fmt.Println("remFile")
	case "addDir":
		fmt.Println("addDir")
	case "remDir":
		fmt.Println("remDir")

	default:
		panic("unknown edit action")
	}
	if err != nil {
		return
	}
	if config.Config.Overrides.Verbose {
		state.meta.Print()
	}
	return state.repo.SaveMeta(state.meta)

}
