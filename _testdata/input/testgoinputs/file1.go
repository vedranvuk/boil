// License lines as doc lines

// Package testgoinputs is a testing package for bast.
package testgoinputs

import "strings"

// Foo defines an interface to foo.
type Foo interface {
	// Bar is a method in Foo.
	Bar(input string) (output string)
}

// Reee defines a struct.
type Reee struct {
	// Tarded is a string field in Reee, with a tag.
	Tarded string `json:"-" db:"tarded"`
}
