// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package bast implements a bastard ast.
package bast

// Bast is a top level struct containign parsed go packages and/or files.
type Bast struct {
	Packages []*Package
}

// Declaration represents a declaration parsed into a File in a Package.
type Declaration interface {
	// GetName returns the element name,
	GetName() string
}

// Package represents a Go package.
type Package struct {
	// Name is the package name, without path.
	Name string
	// Files is a list of files in the package.
	Files []*File
}

// File describes a go source file.
type File struct {
	// Comments are the file comments, grouped by separation, without positions,
	// including docs.
	Comments [][]string
	// Doc is the file doc comment.
	Doc []string
	// Name is the File name, without path.
	Name string
	// Imports is a list of file imports.
	Imports []*Import
	// Declarations is a list of top level declarations in the file.
	Declarations []Declaration
}

// Import represents a package import entry.
type Import struct {
	// Comment is the import comment.
	Comment []string
	// Doc is the import doc.
	Doc []string
	// Name is the import name, possibly empty, "." or some custom name.
	Name string
	// Path is the import path.
	Path string
}

// Interface represents an interface.
type Interface struct {
	// Comment is the interface comment.
	Comment []string
	// Doc is the interface doc comment.
	Doc []string
	// Name is the interface name.
	Name string
	// Methods is a list of methods defined by the interface.
	Methods []*Method
}

// Func represents a func.
type Func struct {
	// Comment is the func comment.
	Comment []string
	// Doc is the func doc comment.
	Doc []string
	// Name is the func name.
	Name string
	//  Arguments is a list of func arguments.
	Arguments []*Pair
	// Returns is a list of func returns.
	Returns []*Pair
}

// Method represents a method.
type Method struct {
	// Func embeds all Func properties.
	Func
	// Receiver is the method receiver.
	Receiver *Pair
}

// Pair represents a key:value/name:type pair..
// It may represent a method receiver, func argument, or result or a struct
// field.
type Pair struct {
	// Name is the left pair part.
	Name string
	// Type is the right pair part.
	Type string
}

// Const represents a constant
type Const struct {
	// Comment is the const comment.
	Comment []string
	// Doc is the const doc comment.
	Doc []string
	// Name is the constant name.
	Name string
	// Type is the const type, empty if undpecified.
	Type string
	// Value is the const value, empty if undpecified.
	Value string
}

// Const represents a constant
type Var struct {
	// Comment is the const comment.
	Comment []string
	// Doc is the const doc comment.
	Doc []string
	// Name is the constant name.
	Name string
	// Type is the const type, empty if undpecified.
	Type string
	// Value is the const value, empty if undpecified.
	Value string
}

// Struct represents a struct type.
type Struct struct {
	// Comment is the struct comment.
	Comment []string
	// Doc is the struct doc comment.
	Doc []string
	// Name is the struct name.
	Name string
	// Fields is a list of struct fields.
	Fields []*Field
}

// Field represents a struct field.
type Field struct {
	// Comment is the field comment.
	Comment []string
	// Doc is the field doc comment.
	Doc []string
	// Name is the field name.
	Name string
	// Type is the field type.
	Type string
	// Tag is the field raw tag string.
	Tag string
}

func (self *Package) GetName() string   { return self.Name }
func (self *File) GetName() string      { return self.Name }
func (self *Import) GetName() string    { return self.Name }
func (self *Interface) GetName() string { return self.Name }
func (self *Func) GetName() string      { return self.Name }
func (self *Method) GetName() string    { return self.Name }
func (self *Pair) GetName() string      { return self.Name }
func (self *Const) GetName() string     { return self.Name }
func (self *Var) GetName() string       { return self.Name }
func (self *Struct) GetName() string    { return self.Name }
func (self *Field) GetName() string     { return self.Name }
