package snap

import (
	"os"

	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Wizard is the Snap command wizard.
type Wizard struct {
	state *state
	*Interrogator
}

// NewWizard returns a new *Wizard.
func NewWizard(state *state) *Wizard {
	return &Wizard{
		state:        state,
		Interrogator: NewInterrogator(os.Stdin, os.Stdout),
	}
}

func (self *Wizard) Execute() (err error) {

	var (
		truth bool
	)

	self.Printf("New template wizard\n\n")

	self.Printf("Template description:\n")
	if self.state.metafile.Description, err = self.AskValue("", ".*"); err != nil {
		return
	}

	self.Printf("Template author name:\n")
	if self.state.metafile.Author.Name, err = self.AskValue(
		self.state.config.Configuration.DefaultAuthor.Name,
		".*",
	); err != nil {
		return err
	}

	self.Printf("Template author email:\n")
	if self.state.metafile.Author.Email, err = self.AskValue(
		self.state.config.Configuration.DefaultAuthor.Email,
		".*",
	); err != nil {
		return
	}

	self.Printf("Template author homepage:\n")
	if self.state.metafile.Author.Homepage, err = self.AskValue(
		self.state.config.Configuration.DefaultAuthor.Homepage,
		".*",
	); err != nil {
		return
	}

	self.Printf("Template version:\n")
	if self.state.metafile.Version, err = self.AskValue(
		"1.0.0",
		".*",
	); err != nil {
		return
	}

	self.Printf("Template URL:\n")
	if self.state.metafile.URL, err = self.AskValue(
		"http://",
		".*",
	); err != nil {
		return
	}

	self.Printf("Would you like to define some Prompts?\n")
	if truth, err = self.AskYesNo(); err != nil {
		return err
	} else if truth {
		if self.state.metafile.Prompts, err = self.definePrompts(); err != nil {
			return
		}
	}

	self.Printf("Would you like to define some Pre-Parse actions?\n")
	if truth, err = self.AskYesNo(); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(&self.state.metafile.Actions.PreParse); err != nil {
			return
		}
	}

	self.Printf("Would you like to define some Pre-Execute actions?\n")
	if truth, err = self.AskYesNo(); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(&self.state.metafile.Actions.PreExecute); err != nil {
			return
		}
	}

	self.Printf("Would you like to define some Post-Execute actions?\n")
	if truth, err = self.AskYesNo(); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(&self.state.metafile.Actions.PostExecute); err != nil {
			return
		}
	}

	self.Printf("Template defined.\n")

	return
}

func (self *Wizard) definePrompts() (result []*boil.Prompt, err error) {

	var (
		prompt *boil.Prompt
		truth  bool
	)

	for {
		prompt = new(boil.Prompt)

		self.Printf("New Prompt:\n")

		self.Printf("Variable name:\n")
		if prompt.Variable, err = self.AskValue("", ".*"); err != nil {
			return
		}

		self.Printf("Variable description:\n")
		if prompt.Description, err = self.AskValue("", ".*"); err != nil {
			return
		}

		self.Printf("Value validation Regular Expression:\n")
		if prompt.RegExp, err = self.AskValue(".*", ".*"); err != nil {
			return
		}

		result = append(result, prompt)

		self.Printf("Would you like to define another Prompt?\n")
		if truth, err = self.AskYesNo(); err != nil {
			return
		} else if truth {
			continue
		}
		break
	}

	return
}

func (self *Wizard) defineActions(actions *boil.Actions) (err error) {

	var (
		action *boil.Action
		truth  bool
	)

	for {
		if action, err = self.defineAction(); err != nil {
			return
		}
		*actions = append(*actions, action)

		self.Printf("Would you like to define another Action?\n")
		if truth, err = self.AskYesNo(); err != nil {
			return err
		} else if truth {
			continue
		}
		break
	}

	return
}

func (self *Wizard) defineAction() (action *boil.Action, err error) {

	action = new(boil.Action)

	var truth bool

	self.Printf("New Action:\n")

	self.Printf("Description:\n")
	if action.Description, err = self.AskValue("", ".*"); err != nil {
		return
	}

	self.Printf("Program:\n")
	if action.Program, err = self.AskValue("", ".*"); err != nil {
		return
	}

	self.Printf("Working directory:\n")
	if action.WorkDir, err = self.AskValue("", ".*"); err != nil {
		return
	}

	self.Printf("Arguments:\n")
	if action.Arguments, err = self.AskList(); err != nil {
		return
	}

	self.Printf("Would you like to define some environment variables?\n")
	if truth, err = self.AskYesNo(); err != nil {
		return
	} else if truth {
		if action.Environment, err = self.defineEnvVariables(); err != nil {
			return
		}
	}

	self.Printf("Don't break the execution if action fails:\n")
	if truth, err = self.AskYesNo(); err != nil {
		return
	} else if truth {
		action.NoFail = true
	}

	return
}

func (self *Wizard) defineEnvVariables() (result map[string]string, err error) {

	var key, val string
	result = make(map[string]string)

	self.Printf("Environment variables:\n")
	for {
		if key, val, err = self.AskVariable(); err != nil {
			return
		}
		if key == "" {
			break
		}
		result[key] = val
	}
	return
}
