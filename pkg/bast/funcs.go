package bast

import (
	"fmt"
	"io"
	"text/template"
)

// FuncMap returns Bast template functions.
func (self Bast) FuncMap() template.FuncMap {
	return template.FuncMap{
		"Declaration":  self.Declaration,
		"VarsOfType":   self.VarsOfType,
		"ConstsOfType": self.ConstsOfType,
	}
}

// Declaration returns a declaration whose name matches from a package named by
// packageName. If packageName is empty declarations is searched in the
// files placeholder package named "command-line-package".
func (self *Bast) Declaration(packageName, name string) (out interface{}) {
	for _, pkg := range self.Packages {
		if packageName == "" && pkg.Name == "command-line-package" {
			return pkg.Declaration(name)
		}
		if pkg.Name == name {
			return pkg.Declaration(name)
		}
	}
	return
}

// ConstsOfType returns all constant declarations from a package named by
// packageName whose type name matches typeName.
func (self Bast) ConstsOfType(packageName, typeName string) (out []Declaration) {
	for _, pkg := range self.Packages {
		out = append(out, pkg.ConstsOfType(typeName)...)
	}
	return
}

// VarsOfType returns all variable declarations from a package named by
// packageName whose type name matches typeName.
func (self Bast) VarsOfType(packageName, typeName string) (out []Declaration) {
	for _, pkg := range self.Packages {
		out = append(out, pkg.VarsOfType(typeName)...)
	}
	return
}

// Print debug prints self to stdout.
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
