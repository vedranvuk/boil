package boil

import (
	"errors"
	"fmt"
	"os"
)

// Wizard is the Snap command wizard.
type Wizard struct {
	config   *Configuration
	metafile *Metafile
	*Interrogator
}

// NewWizard returns a new *Wizard.
func NewWizard(config *Configuration, metafile *Metafile) *Wizard {
	return &Wizard{
		config:       config,
		metafile:     metafile,
		Interrogator: NewInterrogator(os.Stdin, os.Stdout),
	}
}

func (self *Wizard) Execute() (err error) {

	var (
		truth bool
	)

	self.Printf("New template wizard\n\n")

	if err = self.EditInfo(); err != nil {
		return
	}

	self.Printf("Would you like to define some Prompts?\n")
	if truth, err = self.AskYesNo("no"); err != nil {
		return err
	} else if truth {
		if self.metafile.Prompts, err = self.definePrompts(); err != nil {
			return
		}
	}

	self.Printf("Would you like to define some Pre-Parse actions?\n")
	if truth, err = self.AskYesNo("no"); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(&self.metafile.Actions.PreParse); err != nil {
			return
		}
	}

	self.Printf("Would you like to define some Pre-Execute actions?\n")
	if truth, err = self.AskYesNo("no"); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(&self.metafile.Actions.PreExecute); err != nil {
			return
		}
	}

	self.Printf("Would you like to define some Post-Execute actions?\n")
	if truth, err = self.AskYesNo("no"); err != nil {
		return err
	} else if truth {
		if err = self.defineActions(&self.metafile.Actions.PostExecute); err != nil {
			return
		}
	}

	self.Printf("Template defined.\n")

	return
}

func (self *Wizard) definePrompts() (result []*Prompt, err error) {

	var (
		prompt *Prompt
		truth  bool
	)

	for {
		prompt = new(Prompt)

		self.Printf("New Prompt:\n")

		self.Printf("Variable name (enter empty value to abort):\n")
		if prompt.Variable, err = self.AskValue("", ".*"); err != nil {
			return
		}
		if prompt.Variable == "" {
			return nil, nil
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
		if truth, err = self.AskYesNo("no"); err != nil {
			return
		} else if truth {
			continue
		}
		break
	}

	return
}

func (self *Wizard) defineActions(actions *Actions) (err error) {

	var (
		action *Action
		truth  bool
	)

	for {
		if action, err = self.defineAction(); err != nil {
			return
		}
		*actions = append(*actions, action)

		self.Printf("Would you like to define another Action?\n")
		if truth, err = self.AskYesNo("no"); err != nil {
			return err
		} else if truth {
			continue
		}
		break
	}

	return
}

func (self *Wizard) defineAction() (action *Action, err error) {

	action = new(Action)

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
	if truth, err = self.AskYesNo("no"); err != nil {
		return
	} else if truth {
		if action.Environment, err = self.defineEnvVariables(); err != nil {
			return
		}
	}

	self.Printf("Don't break the execution if action fails:\n")
	if truth, err = self.AskYesNo("no"); err != nil {
		return
	} else if truth {
		action.NoFail = true
	}

	return
}

func (self *Wizard) defineEnvVariables() (result map[string]string, err error) {

	var key, val string
	result = make(map[string]string)
	var truth bool

	self.Printf("Environment variables (Enter empty Name to abort):\n")
	for {
		if key, val, err = self.AskVariable(); err != nil {
			return
		}
		if key == "" {
			break
		}
		result[key] = val

		self.Printf("Would you like to define another environment variable?\n")
		if truth, err = self.AskYesNo("no"); err != nil {
			return
		} else if truth {
			continue
		}
		break
	}
	return
}

func (self *Wizard) EditAll() (err error) {
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

func (self *Wizard) EditInfo() (err error) {

	self.Printf("Template description:\n")
	if self.metafile.Description, err = self.AskValue(
		self.metafile.Description,
		".*",
	); err != nil {
		return
	}

	self.Printf("Template author name:\n")
	if self.metafile.Author.Name, err = self.AskValue(
		self.metafile.Author.Name,
		".*",
	); err != nil {
		return err
	}

	self.Printf("Template author email:\n")
	if self.metafile.Author.Email, err = self.AskValue(
		self.metafile.Author.Email,
		".*",
	); err != nil {
		return
	}

	self.Printf("Template author homepage:\n")
	if self.metafile.Author.Homepage, err = self.AskValue(
		self.metafile.Author.Homepage,
		".*",
	); err != nil {
		return
	}

	self.Printf("Template version:\n")
	if self.metafile.Version, err = self.AskValue(
		self.metafile.Version,
		".*",
	); err != nil {
		return
	}

	self.Printf("Template URL:\n")
	if self.metafile.URL, err = self.AskValue(
		self.metafile.URL,
		".*",
	); err != nil {
		return
	}

	return nil
}

func (self *Wizard) EditFiles() error {
	return errors.New("not implemented")
}

func (self *Wizard) EditDirs() error {
	return errors.New("not implemented")
}

func (self *Wizard) editPrompt(prompt *Prompt) (err error) {
	self.Printf("Variable name:\n")
	if prompt.Variable, err = self.AskValue(prompt.Variable, ".*"); err != nil {
		return
	}

	self.Printf("Variable description:\n")
	if prompt.Description, err = self.AskValue(prompt.Description, ".*"); err != nil {
		return
	}

	self.Printf("Value validation Regular Expression:\n")
	if prompt.RegExp, err = self.AskValue(prompt.RegExp, ".*"); err != nil {
		return
	}

	return nil
}

func (self *Wizard) EditPrompts() (err error) {

	// No prompts defined, ask define new
	if len(self.metafile.Prompts) == 0 {
		self.Printf("There are not prompts defined. Would you like to add one?\n")
		var result bool
		if result, err = self.AskYesNo("no"); err != nil {
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
	return self.editPrompt(prompt)
}

func (self *Wizard) EditPreParse() error {
	return errors.New("not implemented")
}

func (self *Wizard) EditPreExec() error {
	return errors.New("not implemented")
}

func (self *Wizard) EditPostExec() error {
	return errors.New("not implemented")
}

func (self *Wizard) EditGroups() error {
	return errors.New("not implemented")
}
