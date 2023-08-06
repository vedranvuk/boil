// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"go/ast"
)

// Package represents a package.
type Package struct {
	Name  string
	Files []*File
}

func PackageFromAST(in *ast.Package) (out *Package) {
	out = new(Package)
	out.Name = in.Name
	for _, file := range in.Files {
		if f := FileFromAST(file); f != nil {
			out.Files = append(out.Files, f)
		}
	}
	return
}

// File describes a go source file.
type File struct {
	Name         string
	Comments     [][]string
	Doc          []string
	Imports      []*Import
	Declarations []interface{}
}

func FileFromAST(in *ast.File) (out *File) {
	out = new(File)
	out.Name = in.Name.Name
	for _, comment := range in.Comments {
		if c := commentGroupToStringSlice(comment); c != nil {
			out.Comments = append(out.Comments, c)
		}
	}
	out.Doc = append(out.Doc, commentGroupToStringSlice(in.Doc)...)
	for _, imprt := range in.Imports {
		if i := ImportFromAST(imprt); i != nil {
			out.Imports = append(out.Imports, i)
		}
	}
	for _, d := range in.Decls {
		if v := loadDecl(d); v != nil {
			out.Declarations = append(out.Declarations, v)
		}
	}
	return
}

func loadDecl(in ast.Decl) (out interface{}) {
	return
}

// Comment represents a comment.
type Comment struct {
	Lines []string
}

func commentGroupToStringSlice(cg *ast.CommentGroup) (result []string) {
	if cg == nil {
		return nil
	}
	for _, entry := range cg.List {
		result = append(result, entry.Text)
	}
	return
}

// Import represents a package import entry.
type Import struct {
	Name    string   // import name or empty string.
	Path    string   // import path.
	Doc     []string // doc comment lines
	Comment []string // comment lines
}

func ImportFromAST(in *ast.ImportSpec) (out *Import) {
	out = new(Import)
	if in.Name != nil {
		out.Name = in.Name.Name
	}
	out.Path = in.Path.Value
	out.Doc = commentGroupToStringSlice(in.Doc)
	out.Comment = commentGroupToStringSlice(in.Comment)
	return
}

// Interface represents an interface.
type Interface struct {
	Name  string
	Funcs []*Func
}

// Func represents a func.
type Func struct {
	Name      string
	Arguments []*Variable
	Returns   []*Variable
}

// Method represents a method.
type Method struct {
	Func
	Receiver string
}

// Variable represents a variable.
type Variable struct {
	Name string
	Type string
}

// Struct represents a struct type.
type Struct struct {
	Name   string
	Fields []*Field
}

// Field represents a struct field.
type Field struct {
	Name string
	Type string
	Tags string
}
