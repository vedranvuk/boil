// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package boil

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"regexp"
	"strings"
	"text/tabwriter"
)

// Interrogator interrogates the user via some reader and writer.
type Interrogator struct {
	rw *bufio.ReadWriter
}

// NewInterrogator returns a new *Interrogator that reads from r and writes to w.
func NewInterrogator(r io.Reader, w io.Writer) *Interrogator {
	return &Interrogator{
		rw: bufio.NewReadWriter(bufio.NewReader(r), bufio.NewWriter(w)),
	}
}

// Printf printfs to self and flushes. Returns an error if one occured.
func (self *Interrogator) Printf(format string, arguments ...any) (err error) {
	if _, err = fmt.Fprintf(self.rw, format, arguments...); err != nil {
		return
	}
	return self.rw.Flush()
}

// Flush flushes self and returns any errors.
func (self *Interrogator) Flush() error { return self.rw.Flush() }

// AskValue asks for a value and returns def if empty string was entered.
// If regex is not empty entered value is matched against it and prompt is
// repeated if the match failed.
// If an error occurs it is returned with an empty result, nil otherwise.
func (self *Interrogator) AskValue(title, def, regex string) (result string, err error) {
	self.Printf("%s [%s]: ", title, def)
	for {
		if result, err = self.rw.ReadString('\n'); err != nil {
			return
		}
		if result = strings.TrimSpace(result); result == "" && def != "" {
			result = def
		}
		if regex != "" {
			var match bool
			if match, err = regexp.MatchString(regex, result); err != nil {
				return "", err
			}
			if !match {
				self.Printf("Invalid value format, try again.\n")
				continue
			}
		}
		if result == "" {
			result = def
		}
		break
	}
	return
}

// AskChoice asks for one of the specified choices and returns it and nil on
// success or an empty result and an error if one occured.
//
// If an empty value is entered the function returns def. There may be no empty
// strings in choices otherwise an error is returned.
//
// The choice string may be a single word that must be repeated on input to
// select a choice. If a non empty value is entered but does not match any of
// the choices the prompt is repeated.
//
// A choice string may also be a tab delimited string where left of the first
// tab is the choice word that must be repeated and right of first tab is the
// short description text displayed next to the choice.
func (self *Interrogator) AskChoice(def string, choices ...string) (result string, err error) {
	for _, choice := range choices {
		if choice == "" {
			return "", errors.New("askchoice: empty string in choices")
		}
	}
PrintChoices:
	var wr = tabwriter.NewWriter(self.rw, 2, 2, 2, 32, 0)
	for _, v := range choices {
		fmt.Fprintf(wr, "%s\n", v)
	}
	if err = wr.Flush(); err != nil {
		return
	}
	if err = self.Flush(); err != nil {
		return
	}
	self.Printf("Choose a value [%s]: ", def)
	if err = wr.Flush(); err != nil {
		return
	}
	if err = self.rw.Flush(); err != nil {
		return
	}
Prompt:
	for {
		if result, err = self.rw.ReadString('\n'); err != nil {
			return
		}
		result, _, _ = strings.Cut(result, "\t")
		if result = strings.TrimSpace(result); result == "" {
			return def, nil
		}
		for _, choice := range choices {
			choice, _, _ = strings.Cut(choice, "\t")
			if result == choice {
				break Prompt
			}
		}
		self.Printf("Invalid choice, try again.\n")
		goto PrintChoices
	}
	return
}

// AskYesNo asks for a choice between "yes" or a "no" using AskChoice.
// If an empty string is entered the function returns def.
// If a word other than "yes" and "no" is entered the prompt is repeated.
// The value of def must be "yes" or "no" or an error is returned.
// If an error occurs returns empty result and an error or nil otherwise.
func (self *Interrogator) AskYesNo(def bool) (result bool, err error) {
	var response string
	var d string
	if def {
		d = "yes"
	} else {
		d = "no"
	}
	if response, err = self.AskChoice(d, "yes", "no"); err != nil {
		return
	}
	return response == "yes", nil
}

// AskList asks for a list of values by repeatedly asking for a value until an
// empty string is entered then returns the result and a nil error or an empty
// result and an error if one occured.
func (self *Interrogator) AskList() (result []string, err error) {
	self.Printf("Define a list of values (enter empty value to finish).\n")
	var val string
	for {
		if val, err = self.AskValue("Value", "", ".*"); err != nil {
			return
		}
		if val = strings.TrimSpace(val); val == "" {
			break
		}
		result = append(result, val)
	}
	return
}

// AskVariable asks for a key=value pair and returns them with a nil error.
// If an empty key is entered the function aborts and returns empty key and
// value and a nil error. Caller should check validity of returned values,
// If any other error occurs returns empty key and value and the occured error.
func (self *Interrogator) AskVariable() (key, value string, err error) {
	self.Printf("Define a variable.\n")
	self.Printf("Name:\n")
	if key, err = self.AskValue("Name", "", ".*"); err != nil {
		return "", "", err
	}
	self.Printf("Value:\n")
	if value, err = self.AskValue("Value", "", ".*"); err != nil {
		return "", "", err
	}
	return
}
