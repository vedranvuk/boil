// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package bast implements a bastard ast.
package bast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"os"
	"text/template"
)

// Load loads bast of inputs which can be module paths, absolute or relative
// paths to go files or packages. If no inputs are given Load returns an empty
// bast. if an error occurs it is returned.
func Load(inputs ...string) (bast *Bast, err error) {

	bast = new(Bast)

	const parseMode = parser.ParseComments | parser.DeclarationErrors | parser.AllErrors // | parser.Trace

	var (
		fp *Package
		fi os.FileInfo
		ff = token.NewFileSet()
	)

	for _, input := range inputs {
		if fi, err = os.Stat(input); err != nil {
			err = fmt.Errorf("stat input: %w", err)
			return
		}
		// Load complete package...
		if fi.IsDir() {
			var (
				fs   = token.NewFileSet()
				pkgs map[string]*ast.Package
			)
			if pkgs, err = parser.ParseDir(fs, input, nil, parseMode); err != nil {
				return
			}
			for _, pkg := range pkgs {
				appendPackage(fs, pkg, &bast.Packages)
			}
			continue
		}
		// ... or load file into placeholder root package.
		if fp == nil {
			fp = new(Package)
			fp.Name = "command-line-package"
		}
		var f *ast.File
		if f, err = parser.ParseFile(ff, input, nil, parseMode); err != nil {
			return
		}
		appendFile(ff, f, &fp.Files)
	}

	// Add placeholder package to parsed packages.
	if fp != nil {
		bast.Packages = append(bast.Packages, fp)
	}

	return
}

// Bast is a top level struct for accessing processed go files.
type Bast struct {
	Packages []*Package
}

func (self *Bast) Print(w io.Writer) {
	for _, pkg := range self.Packages {
		fmt.Printf("Package %s\n", pkg.Name)
		for _, file := range pkg.Files {
			fmt.Printf("  File: %s\n", file.Name)
			for _, decl := range file.Declarations {
				switch d := decl.(type) {
				case *Struct:
					fmt.Printf("    Struct %s\n", d.Name)
					for _, field := range d.Fields {
						fmt.Printf("      Field %s (%s)\n", field.Name, field.Type)
					}
				case *Interface:
					fmt.Printf("    Interface %s\n", d.Name)
					for _, method := range d.Methods {
						fmt.Printf("      Method %s\n", method.Name)
					}
				}
			}
		}
	}
}

// FuncMap returns Bast template functions.
func (self Bast) FuncMap() template.FuncMap {
	return template.FuncMap{
		"declaration": self.Declaration,
	}
}

// Declaration returns a top level declaration with the specified name.
// Template must know the type of the returned declaration.
func (self Bast) Declaration(name string) (result interface{}) {
	return
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
	Declarations []interface{}
}

// Import represents a package import entry.
type Import struct {
	// Comment is the import comment.
	Comment []string
	// Doc is the import doc.
	Doc []string
	// Name is the custom import name, possibly empty.
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

func appendPackage(fs *token.FileSet, in *ast.Package, out *[]*Package) {
	var val = new(Package)
	val.Name = in.Name

	for _, file := range in.Files {
		appendFile(fs, file, &val.Files)
	}

	return
}

func appendFile(fs *token.FileSet, in *ast.File, out *[]*File) {
	var val = new(File)
	val.Name = in.Name.Name

	var cg []string
	for _, comment := range in.Comments {
		appendCommentGroup(comment, &cg)
		val.Comments = append(val.Comments, cg)
	}

	appendCommentGroup(in.Doc, &val.Doc)

	for _, imprt := range in.Imports {
		appendImportSpec(imprt, &val.Imports)
	}

	for _, d := range in.Decls {
		appendDeclaration(fs, d.(ast.Node), &val.Declarations)
	}

	*out = append(*out, val)

	return
}

func appendDeclaration(fs *token.FileSet, in ast.Node, out *[]interface{}) {
	switch n := in.(type) {
	case *ast.GenDecl:
		if n.Tok != token.TYPE {
			return
		}
		for _, spec := range n.Specs {

			var ts, ok = spec.(*ast.TypeSpec)
			if !ok || ts.Assign != token.NoPos {
				continue
			}

			switch ts.Type.(type) {
			case *ast.InterfaceType:
				appendInterface(ts, out)
			case *ast.StructType:
				appendStruct(ts, out)
			}
		}
	}
	return
}

func appendCommentGroup(in *ast.CommentGroup, out *[]string) {
	if in == nil {
		return
	}

	for _, entry := range in.List {
		*out = append(*out, entry.Text)
	}

	return
}

func appendImportSpec(in *ast.ImportSpec, out *[]*Import) {
	var val = new(Import)
	if in.Name != nil {
		val.Name = in.Name.Name
	}
	val.Path = in.Path.Value
	appendCommentGroup(in.Doc, &val.Doc)
	appendCommentGroup(in.Comment, &val.Comment)
	return
}

func appendInterface(in *ast.TypeSpec, out *[]interface{}) {
	var it, ok = in.Type.(*ast.InterfaceType)
	if !ok {
		return
	}
	var val = new(Interface)
	appendCommentGroup(in.Comment, &val.Comment)
	appendCommentGroup(in.Doc, &val.Doc)
	val.Name = in.Name.Name
	for _, method := range it.Methods.List {
		appendMethod(method, &val.Methods)
	}
	*out = append(*out, val)
	return
}

func appendStruct(in *ast.TypeSpec, out *[]interface{}) {
	var st, ok = in.Type.(*ast.StructType)
	if !ok {
		return
	}
	var val = new(Struct)
	appendCommentGroup(in.Comment, &val.Comment)
	appendCommentGroup(in.Doc, &val.Doc)
	val.Name = in.Name.Name
	for _, field := range st.Fields.List {
		appendField(field, &val.Fields)
	}
	*out = append(*out, val)
	return
}

func appendMethod(in *ast.Field, out *[]*Method) {
	var val = new(Method)
	if len(in.Names) > 0 {
		val.Name = in.Names[0].Name
	}
	appendCommentGroup(in.Comment, &val.Comment)
	appendCommentGroup(in.Doc, &val.Doc)
	var ft, ok = in.Type.(*ast.FuncType)
	if !ok {
		return
	}

	if ft.TypeParams != nil {
		val.Receiver = &Pair{
			Name: ft.TypeParams.List[0].Names[0].Name,
			Type: ft.TypeParams.List[0].Type.(*ast.Ident).Name,
		}
	}

	if ft.Params != nil {
		var arg = new(Pair)
		for _, param := range ft.Params.List {
			arg.Name = param.Names[0].Name
			arg.Type = param.Type.(*ast.Ident).Name
			val.Arguments = append(val.Arguments, arg)
		}
	}

	if ft.Results != nil {
		var arg = new(Pair)
		for _, param := range ft.Results.List {
			arg.Name = param.Names[0].Name
			arg.Type = param.Type.(*ast.Ident).Name
			val.Returns = append(val.Returns, arg)
		}
	}

	return
}

func appendField(in *ast.Field, out *[]*Field) {
	var val = new(Field)
	if len(in.Names) > 0 {
		val.Name = in.Names[0].Name
	}
	appendCommentGroup(in.Comment, &val.Comment)
	appendCommentGroup(in.Doc, &val.Doc)

	val.Type = in.Type.(*ast.Ident).Name
	val.Tag = in.Tag.Value

	return
}
