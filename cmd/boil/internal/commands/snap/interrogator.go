package snap

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Interrogator interrigates, asks questions.
type Interrogator struct {
	rw *bufio.ReadWriter
}

// NewInterrogator returns a new *Interrogator that reads from r and writes to w.
func NewInterrogator(r io.Reader, w io.Writer) *Interrogator {
	return &Interrogator{
		rw: bufio.NewReadWriter(bufio.NewReader(r), bufio.NewWriter(w)),
	}
}

// Flush flushes self.
func (self *Interrogator) Flush() error { return self.rw.Flush() }

// AskValue asks for a value on a new line for something named with name.
// Returns def in an empty string was entered.
// If regex is not empty entered value is matched against it and question
// repeated if the match failed.
// If an error occurs it is returned with an empty result, nil otherwise.
func (self *Interrogator) AskValue(def, regex string) (result string, err error) {
	fmt.Fprintf(self.rw, "Enter value (Default: '%s'):\n", def)
	for {
		if result, err = self.rw.ReadString('\n'); err != nil {
			return
		}
		result = strings.TrimSpace(result)
		if regex != "" {
			var match bool
			if match, err = regexp.MatchString(regex, result); err != nil {
				return "", err
			}
			if !match {
				continue
			}
			break
		}
	}
	return
}

// AskChoice asks for one of the specified choices on a new line.
// Returns def if an empty string was entered.
// Repeats the question until one of the choices is given.
// If an error occurs it is returned with an empty result, nil otherwise.
func (self *Interrogator) AskChoice(def string, choices ...string) (result string, err error) {
PrintChoices:
	fmt.Fprintf(self.rw, "Choose a value (Default: '%s'):\n", def)
	for _, v := range choices {
		fmt.Fprintf(self.rw, "%s\n", v)
	}
Prompt:
	for {
		fmt.Fprintf(self.rw, "\nEnter value:\n")
		if result, err = self.rw.ReadString('\n'); err != nil {
			return
		}
		if result = strings.TrimSpace(result); result == "" {
			result = def
		}
		for _, choice := range choices {
			if result == choice {
				break Prompt
			}
		}
		fmt.Fprintf(self.rw, "Try again.\n\n")
		goto PrintChoices
	}
	return
}

// AskYesNo asks for a "yes" or a "no".
func (self *Interrogator) AskYesNo() (result bool, err error) {

	var response string

	if response, err = self.AskChoice("yes", "yes", "no"); err != nil {
		return
	}

	return response == "yes", nil
}

// AskList asks for a list of values.
// Prompting stops on first blank value entered.
func (self *Interrogator) AskList() (result []string, err error) {

	var val string

	fmt.Fprintf(self.rw, "Define a list of values. Enter an empty string to finish.\n")
	for {
		if val, err = self.AskValue(".*", ""); err != nil {
			return
		}
		if val = strings.TrimSpace(val); val == "" {
			break
		}
		result = append(result, val)
	}

	return
}

func (self *Interrogator) AskVariable() (key, value string, err error) {

	fmt.Fprintf(self.rw, "Define a variable. Enter an empty string for Name to finish.\n")

	fmt.Fprintf(self.rw, "Name:\n")
	if key, err = self.AskValue("", ".*"); err != nil {
		return "", "", nil
	}

	fmt.Fprintf(self.rw, "Value:\n")
	if value, err = self.AskValue("", ".*"); err != nil {
		return
	}

	return
}
