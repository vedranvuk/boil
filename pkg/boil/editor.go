// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package boil

import (
	"fmt"
	"os"
)

// Editor is a metafile editor that defines or edits metafiles using stdio for
// user interaction.
type Editor struct {
	config   *Config
	metafile *Metafile
	*Interrogator
}

// NewEditor returns a new metafile *Editor configured by config.
func NewEditor(config *Config, metafile *Metafile) *Editor {
	return &Editor{
		config:       config,
		metafile:     metafile,
		Interrogator: NewInterrogator(os.Stdin, os.Stdout),
	}
}

// Wizard executes a wizard that completely defines the loaded metafile.
func (self *Editor) Wizard() (err error) {

	var truth bool

	self.Printf("New Template Wizard\n\n")
	if err = self.EditInfo(); err != nil {
		return
	}
	self.Printf("Define a new prompt?\n")
	if truth, err = self.AskYesNo(false); err != nil {
		return err
	} else if truth {
		if self.metafile.Prompts, err = self.definePrompts(); err != nil {
			return
		}
	}
	self.Printf("Define a new Pre-Parse action?\n")
	if truth, err = self.AskYesNo(false); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(&self.metafile.Actions.PreParse); err != nil {
			return
		}
	}
	self.Printf("Define a new Pre-Execute action?\n")
	if truth, err = self.AskYesNo(false); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(&self.metafile.Actions.PreExecute); err != nil {
			return
		}
	}
	self.Printf("Define a new Post-Execute action?\n")
	if truth, err = self.AskYesNo(false); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(&self.metafile.Actions.PostExecute); err != nil {
			return
		}
	}
	self.Printf("Template defined.\n")
	return
}

func (self *Editor) definePrompts() (result []*Prompt, err error) {

	var (
		prompt *Prompt
		truth  bool
	)

	for {
		self.Printf("New Prompt\n")
		prompt = new(Prompt)

		if prompt.Variable, err = self.AskValue("Variable", "", ".*"); err != nil {
			return
		}
		if prompt.Variable == "" {
			return nil, nil
		}
		if prompt.Description, err = self.AskValue("Description", "", ".*"); err != nil {
			return
		}
		if prompt.RegExp, err = self.AskValue("Regular Expression", ".*", ".*"); err != nil {
			return
		}
		self.Printf("Is optional (don't raise error on empty value)?\n")
		if prompt.Optional, err = self.AskYesNo(false); err != nil {
			return
		}
		result = append(result, prompt)

		self.Printf("Define another Prompt?\n")
		if truth, err = self.AskYesNo(false); err != nil {
			return
		} else if truth {
			continue
		}
		break
	}

	return
}

func (self *Editor) defineActions(actions *Actions) (err error) {

	var (
		action *Action
		truth  bool
	)

	for {
		if action, err = self.defineAction(); err != nil {
			return
		}
		*actions = append(*actions, action)

		self.Printf("Define another Action?\n")
		if truth, err = self.AskYesNo(false); err != nil {
			return err
		} else if truth {
			continue
		}
		break
	}

	return
}

func (self *Editor) defineAction() (action *Action, err error) {

	var truth bool
	action = new(Action)
	self.Printf("New Action\n")

	if action.Description, err = self.AskValue("Description", "", ".*"); err != nil {
		return
	}
	if action.Program, err = self.AskValue("Program", "", ".*"); err != nil {
		return
	}
	if action.WorkDir, err = self.AskValue("Working directory", "", ".*"); err != nil {
		return
	}

	self.Printf("Arguments\n")
	if action.Arguments, err = self.AskList(); err != nil {
		return
	}

	self.Printf("Define an environment variable?\n")
	if truth, err = self.AskYesNo(false); err != nil {
		return
	} else if truth {
		if action.Environment, err = self.defineEnvVariables(); err != nil {
			return
		}
	}

	self.Printf("Don't break execution if action fails?\n")
	if truth, err = self.AskYesNo(false); err != nil {
		return
	} else if truth {
		action.NoFail = true
	}

	return
}

func (self *Editor) defineEnvVariables() (result map[string]string, err error) {

	var key, val string
	result = make(map[string]string)
	var truth bool

	self.Printf("Environment Variables (Enter empty Name to abort)\n")
	for {
		if key, val, err = self.AskVariable(); err != nil {
			return
		}
		if key == "" {
			break
		}
		result[key] = val

		self.Printf("Define another environment variable?\n")
		if truth, err = self.AskYesNo(false); err != nil {
			return
		} else if truth {
			continue
		}
		break
	}
	return
}

func (self *Editor) EditAll() (err error) {
	if err = self.EditInfo(); err != nil {
		return
	}
	if err = self.EditFiles(); err != nil {
		return
	}
	if err = self.EditDirs(); err != nil {
		return
	}
	if err = self.EditPrompts(); err != nil {
		return
	}
	if err = self.EditPreParse(); err != nil {
		return
	}
	if err = self.EditPreExec(); err != nil {
		return
	}
	if err = self.EditPostExec(); err != nil {
		return
	}
	if err = self.EditGroups(); err != nil {
		return
	}
	return nil
}

func (self *Editor) EditInfo() (err error) {
	if self.metafile.Description, err = self.AskValue(
		"Description", self.metafile.Description, ".*"); err != nil {
		return
	}
	if self.metafile.Version, err = self.AskValue(
		"Version", self.metafile.Version, ".*"); err != nil {
		return
	}
	if self.metafile.URL, err = self.AskValue(
		"URL", self.metafile.URL, ".*"); err != nil {
		return
	}
	if self.metafile.Author.Name, err = self.AskValue(
		"Author name", self.metafile.Author.Name, ".*"); err != nil {
		return err
	}
	if self.metafile.Author.Email, err = self.AskValue(
		"Author email", self.metafile.Author.Email, ".*"); err != nil {
		return
	}
	if self.metafile.Author.Homepage, err = self.AskValue(
		"Author homepage", self.metafile.Author.Homepage, ".*"); err != nil {
		return
	}
	return nil
}

func (self *Editor) EditFiles() error {
	// TODO: Implement Editor.EditFiles.
	return nil
}

func (self *Editor) EditDirs() error {
	// TODO: Implement Editor.EditDirs.
	return nil
}

func (self *Editor) EditPrompt(prompt *Prompt) (err error) {
	if prompt.Variable, err = self.AskValue(
		"Variable", prompt.Variable, ".*"); err != nil {
		return
	}
	if prompt.Description, err = self.AskValue(
		"Description", prompt.Description, ".*"); err != nil {
		return
	}
	if prompt.RegExp, err = self.AskValue(
		"Regular Expression", prompt.RegExp, ".*"); err != nil {
		return
	}
	self.Printf("Is optional (don't raise error on empty value entered)?\n")
	if prompt.Optional, err = self.AskYesNo(false); err != nil {
		return
	}
	return nil
}

func (self *Editor) EditPrompts() (err error) {

	// No prompts defined, ask define new
	if len(self.metafile.Prompts) == 0 {
		self.Printf("There are not prompts defined. Would you like to add one?\n")
		var result bool
		if result, err = self.AskYesNo(false); err != nil {
			return
		}
		if !result {
			return nil
		}
		var prompts []*Prompt
		if prompts, err = self.definePrompts(); err != nil {
			return
		}
		self.metafile.Prompts = append(self.metafile.Prompts, prompts...)
	}

	// Edit existing
	if len(self.metafile.Prompts) == 0 {
		return nil
	}
	self.Printf("Select prompt to edit (empty value to stop):\n")

	var (
		prompt   *Prompt
		choices  []string
		variable string
	)

	for _, prompt = range self.metafile.Prompts {
		choices = append(choices, fmt.Sprintf("%s\t%s", prompt.Variable, prompt.Description))
	}
	if variable, err = self.AskChoice("", choices...); err != nil {
		return
	}
	if variable == "" {
		return nil
	}
	if prompt = self.metafile.Prompts.FindByVariable(variable); prompt == nil {
		panic("prompt not found")
	}
	return self.EditPrompt(prompt)
}

func (self *Editor) EditPreParse() error {
	// TODO: Implement Editor.EditPreParse.
	return nil
}

func (self *Editor) EditPreExec() error {
	// TODO: Implement Editor.EditPreExec.
	return nil
}

func (self *Editor) EditPostExec() error {
	// TODO: Implement Editor.EditPostExec.
	return nil
}

func (self *Editor) EditGroups() error {
	// TODO: Implement Editor.EditGroups.
	return nil
}
