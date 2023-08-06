// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package bast implements a bastard ast.
package bast

import (
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"text/template"

	"golang.org/x/tools/go/packages"
)

// Load loads bast of inputs which can be module paths, absolute or relative
// paths to go files or packages. If no inputs are given Load returns an empty
// bast. if an error occurs it is returned.
func Load(inputs ...string) (bast *Bast, err error) {

	bast = new(Bast)

	const parseMode = parser.ParseComments | parser.Trace | parser.DeclarationErrors | parser.AllErrors

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
		if fi.IsDir() {
			var (
				fs   = token.NewFileSet()
				pkgs map[string]*ast.Package
			)
			if pkgs, err = parser.ParseDir(fs, input, nil, parseMode); err != nil {
				return
			}
			for _, p := range pkgs {
				bast.Packages = append(bast.Packages, PackageFromAST(p))
			}
		} else {
			if fp == nil {
				fp = new(Package)
				fp.Name = "command-line-package"
			}
			var f *ast.File
			if f, err = parser.ParseFile(ff, input, nil, parseMode); err != nil {
				return
			}
			fp.Files = append(fp.Files, FileFromAST(f))
		}
	}
	if fp != nil {
		bast.Packages = append(bast.Packages, fp)
	}

	var (
		fileSet = token.NewFileSet()
		pkgs    []*packages.Package
	)

	if pkgs, err = packages.Load(&packages.Config{
		Mode: packages.NeedName | packages.NeedFiles | packages.NeedSyntax,
		Fset: fileSet,
	}, inputs...); err != nil {
		return
	}
	if packages.PrintErrors(pkgs) > 0 {
		err = errors.New("errors occured")
		return
	}

	return
}

// Bast is a top level struct for accessing processed go files.
type Bast struct {
	Packages []*Package
}

// FuncMap returns Bast template functions.
func (self Bast) FuncMap() template.FuncMap {
	return template.FuncMap{
		"": "",
	}
}
