// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

func (self *Package) Declaration(name string) (out Declaration) {
	for _, file := range self.Files {
		for _, decl := range file.Declarations {
			if decl.GetName() == name {
				return decl
			}
		}
	}
	return
}

func (self *Package) ConstsOfType(name string) (out []Declaration) {
	for _, file := range self.Files {
		for _, decl := range file.Declarations {
			var c, ok = decl.(*Const)
			if !ok {
				continue
			}
			if c.Type == name {
				out = append(out, decl)
			}
		}
	}
	return
}

func (self *Package) VarsOfType(name string) (out []Declaration) {
	for _, file := range self.Files {
		for _, decl := range file.Declarations {
			var c, ok = decl.(*Var)
			if !ok {
				continue
			}
			if c.Type == name {
				out = append(out, decl)
			}
		}
	}
	return
}