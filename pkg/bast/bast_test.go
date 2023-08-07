package bast

import "testing"

func TestConstDecl(t *testing.T) {
	const src = `package consttest

const Foo string = "Bar"
`

	var (
		data *Bast
		err  error
	)

	if data, err = ParseSrc(src); err != nil || data == nil {
		t.Fatal(err)
	}

	if data.Declaration("", "Foo").(*Const).Value != "Bar" {
		t.Fatalf("Const decl failed.")
	}
}

func TestVarDecl(t *testing.T) {
	const src = `package consttest

var Foo string = "Bar"
`

	var (
		data *Bast
		err  error
	)

	if data, err = ParseSrc(src); err != nil || data == nil {
		t.Fatal(err)
	}

	if data.Declaration("", "Foo").(*Var).Value != "Bar" {
		t.Fatalf("Const decl failed.")
	}
}
