package snap

import (
	"fmt"
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

	var truth bool

	fmt.Fprintf(self.rw, "Template description:\n")
	if self.state.metafile.Description, err = self.AskValue(".*", ""); err != nil {
		return
	}

	fmt.Fprintf(self.rw, "Template author name:\n")
	if self.state.metafile.Author.Name, err = self.AskValue(
		".*",
		self.state.config.Configuration.DefaultAuthor.Name,
	); err != nil {
		return err
	}

	fmt.Fprintf(self.rw, "Template author email:\n")
	if self.state.metafile.Author.Email, err = self.AskValue(
		".*",
		self.state.config.Configuration.DefaultAuthor.Email,
	); err != nil {
		return
	}

	fmt.Fprintf(self.rw, "Template author homepage:\n")
	if self.state.metafile.Author.Homepage, err = self.AskValue(
		".*",
		self.state.config.Configuration.DefaultAuthor.Homepage,
	); err != nil {
		return
	}

	fmt.Fprintf(self.rw, "Template version:\n")
	if self.state.metafile.Version, err = self.AskValue(
		".*",
		"1.0.0",
	); err != nil {
		return
	}

	fmt.Fprintf(self.rw, "Template URL:\n")
	if self.state.metafile.URL, err = self.AskValue(
		".*",
		"http://",
	); err != nil {
		return
	}

	fmt.Fprintf(self.rw, "Would you like to define some Prompts?\n")
	if truth, err = self.AskYesNo(); err != nil {
		return err
	} else if truth {
		if self.state.metafile.Prompts, err = self.definePrompts(); err != nil {
			return
		}
	}

	fmt.Fprintf(self.rw, "Would you like to define some Pre-Parse actions?\n")
	if truth, err = self.AskYesNo(); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(self.state.metafile.Actions.PreParse); err != nil {
			return
		}
	}

	fmt.Fprintf(self.rw, "Would you like to define some Pre-Execute actions?\n")
	if truth, err = self.AskYesNo(); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(self.state.metafile.Actions.PreExecute); err != nil {
			return
		}
	}

	fmt.Fprintf(self.rw, "Would you like to define some Post-Execute actions?\n")
	if truth, err = self.AskYesNo(); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(self.state.metafile.Actions.PostExecute); err != nil {
			return
		}
	}

	return
}

func (self *Wizard) definePrompts() (result []*boil.Prompt, err error) {

	var (
		prompt *boil.Prompt
		truth  bool
	)

	for {
		prompt = new(boil.Prompt)

		fmt.Fprintf(self.rw, "Enter Variable name:\n")
		if prompt.Variable, err = self.AskValue("", ".*"); err != nil {
			return
		}

		fmt.Fprintf(self.rw, "Enter Variable description:\n")
		if prompt.Description, err = self.AskValue("", ".*"); err != nil {
			return
		}

		fmt.Fprintf(self.rw, "Enter regular expression to use for checking value:\n")
		if prompt.RegExp, err = self.AskValue(".*", ".*"); err != nil {
			return
		}

		result = append(result, prompt)

		fmt.Fprintf(self.rw, "Would you like to define another Prompt?\n")
		if truth, err = self.AskYesNo(); err != nil {
			return
		} else if truth {
			continue
		}
		break
	}

	return
}

func (self *Wizard) defineActions(actions boil.Actions) (err error) {

	var (
		action *boil.Action
		truth  bool
	)

	for {
		if action, err = self.defineAction(); err != nil {
			return
		}
		actions = append(actions, action)

		fmt.Fprintf(self.rw, "Would you like to define another Action?\n")
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

	fmt.Printf("Define a new Action:\n")

	fmt.Fprintf(self.rw, "Description:\n")
	if action.Description, err = self.AskValue("", ".*"); err != nil {
		return
	}

	fmt.Fprintf(self.rw, "Program:\n")
	if action.Program, err = self.AskValue("", ".*"); err != nil {
		return
	}

	fmt.Fprintf(self.rw, "Working directory:\n")
	if action.WorkDir, err = self.AskValue("", ".*"); err != nil {
		return
	}
	
	fmt.Fprintf(self.rw, "Description:\n")
	if action.Arguments, err = self.AskList(); err != nil {
		return
	}

	fmt.Fprintf(self.rw, "Would you like to define some environment variables?\n")
	if truth, err = self.AskYesNo(); err != nil {
		return
	} else if truth {
		if action.Environment, err = self.defineEnvVariables(); err != nil {
			return
		}
	}

	fmt.Fprintf(self.rw, "Environment variables:\n")
	if action.Environment, err = self.defineEnvVariables(); err != nil {
		return
	}

	fmt.Fprintf(self.rw, "Don't break the execution if action fails:\n")
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

	fmt.Fprintf(self.rw, "Environment variables:\n")
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
