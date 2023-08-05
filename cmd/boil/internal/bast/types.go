package bast

type Package struct {
	Name  string
	Files []File
}

// File describes a go source file.
type File struct {
	Name         string
	Comments     []Comment
	Doc          []string
	Imports      []Import
	Declarations []interface{}
}

type Comment struct {
	Lines []string
}

type Import struct {
	Name    string   // custom import name or empty string.
	Doc     []string // doc comment lines
	Comment []string // comment lines
}

type Interface struct {
	Name    string
	Methods []*Method
}

type Method struct {
	Name      string
	Arguments []*Variable
	Returns   []*Variable
}

type Variable struct {
	Name string
	Type string
}

type Struct struct {
	Name   string
	Fields []Field
}

type Field struct {
	Name string
	Type string
	Tags string
}
