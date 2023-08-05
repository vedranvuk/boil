// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"go/token"

	"golang.org/x/tools/go/packages"
)

// Bast is a top level struct for accessing processed go files.
type Bast struct {
}

// Load loads bast of inputs which can be module paths, absolute or relative
// paths to go files or packages. If no inputs are given Load returns an empty
// bast. if an error occurs it is returned.
func Load(inputs ...string) (bast *Bast, err error) {

	bast = new(Bast)

	const (
		loadMode = packages.NeedName | packages.NeedFiles |
			packages.NeedTypes | packages.NeedSyntax
	)

	var (
		fileSet = token.NewFileSet()
		pkgs    []*packages.Package
	)

	if pkgs, err = packages.Load(&packages.Config{
		Mode: loadMode,
		Fset: fileSet,
	}, inputs...); err != nil {
		return
	}

	for _, pkg := range pkgs {
		_ = pkg
	}
	return
}
