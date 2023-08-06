// License lines as doc lines

// Package testgoinputs is a testing package for bast.
package testgoinputs

import "strings"

type MyInterface interface {
	MyMethod(input string) (output string)
}

type MyStruct struct {
	Name string
	Age  int
}

type (
	MyStruct2 struct {
		Name string
	}
	MyStruct3 struct {
		Name string
		Sub  struct {
			Surname string
		}
	}
)

func TestFunc(in string) (out stirng) { return strings.ToLower(in) }

type TestStruct struct{}

func (self TestStruct) Method() {}
func (self *TestStruct) PointerMethod() {}
