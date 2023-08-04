// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package exec

import (
	"github.com/vedranvuk/boil/cmd/boil/internal/boil"
)

// Data is the top level data structure passed to a Template file.
type Data struct {
	// Vars is a collection of system variables always present on template
	// execution, generated from environment.
	Vars boil.Variables
}

// NeData returns new *Data initialized with standard variables or an error if
// one occurs. If merge is not nil it is added to the resulting Data.
func NewData() *Data {
	return &Data{
		Vars: make(boil.Variables),
	}
}

// ReplaceAll replaces all known variable placeholders in input string with
// actual values and returns it.
func (self *Data) ReplaceAll(in string) (out string) {
	return self.Vars.ReplacePlaceholders(in)
}
