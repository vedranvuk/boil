package snap

import (
	"bufio"
	"fmt"
	"io"
	"regexp"
	"strings"
)

// Interrogator interrigates the user via some reader and writer.
type Interrogator struct {
	rw *bufio.ReadWriter
}

// NewInterrogator returns a new *Interrogator that reads from r and writes to w.
func NewInterrogator(r io.Reader, w io.Writer) *Interrogator {
	return &Interrogator{
		rw: bufio.NewReadWriter(bufio.NewReader(r), bufio.NewWriter(w)),
	}
}

// Printf printfs to self.
func (self *Interrogator) Printf(format string, arguments ...any) (err error) {
	if _, err = fmt.Fprintf(self.rw, format, arguments...); err != nil {
		return
	}
	return self.rw.Flush()
}

// Flush flushes self.
func (self *Interrogator) Flush() error { return self.rw.Flush() }

// AskValue asks for a value on a new line for something named with name.
// Returns def in an empty string was entered.
// If regex is not empty entered value is matched against it and question
// repeated if the match failed.
// If an error occurs it is returned with an empty result, nil otherwise.
func (self *Interrogator) AskValue(def, regex string) (result string, err error) {
	self.Printf("Enter value (Default: '%s'):\n", def)
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
				self.Printf("Invalid value format. Try again\n")
				continue
			}
		}
		break
	}
	return
}

// AskChoice asks for one of the specified choices on a new line.
// Returns def if an empty string was entered.
// Repeats the question until one of the choices is given.
// If an error occurs it is returned with an empty result, nil otherwise.
func (self *Interrogator) AskChoice(def string, choices ...string) (result string, err error) {
PrintChoices:
	self.Printf("Choose a value (Default: '%s'):\n", def)
	for _, v := range choices {
		self.Printf("* %s\n", v)
	}
Prompt:
	for {
		self.Printf("\nEnter value:\n")
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
		self.Printf("Try again.\n\n")
		goto PrintChoices
	}
	return
}

// AskYesNo asks for a "yes" or a "no".
func (self *Interrogator) AskYesNo() (result bool, err error) {

	var response string

	if response, err = self.AskChoice("no", "yes", "no"); err != nil {
		return
	}

	return response == "yes", nil
}

// AskList asks for a list of values.
// Prompting stops on first blank value entered.
func (self *Interrogator) AskList() (result []string, err error) {

	var val string

	self.Printf("Define a list of values. Enter an empty string to finish.\n")
	for {
		if val, err = self.AskValue("", ".*"); err != nil {
			return
		}
		if val = strings.TrimSpace(val); val == "" {
			break
		}
		result = append(result, val)
	}

	return
}

// AskVariable asks for a key=value pair.
// Prompt is aborted if empty name entered, returns empty keyval and nil.
func (self *Interrogator) AskVariable() (key, value string, err error) {

	self.Printf("Define a variable.\n")

	self.Printf("Name (Enter empty string to abort):\n")
	if key, err = self.AskValue("", ".*"); err != nil {
		return "", "", err
	}

	if key == "" {
		return "", "", nil
	}

	self.Printf("Value:\n")
	if value, err = self.AskValue("", ".*"); err != nil {
		return "", "", err
	}

	return
}
