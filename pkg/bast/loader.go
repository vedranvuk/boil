// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

package bast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"strconv"
)

// Load loads bast of inputs which can be module paths, absolute or relative
// paths to go files or packages. If no inputs are given Load returns an empty
// bast.
//
// Inputs that point to files, i.e. are outside of a package are put into a
// placeholder package named "command-line-package" which mirrors how
// "golang.org/x/tools/go/packages" names it.
//
// If an error occurs it is returned.
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

// ParseSrc returns a Bast of input source src or an error if one occurs.
func ParseSrc(src string) (bast *Bast, err error) {
	bast = new(Bast)
	var (
		pkg  = new(Package)
		fset = token.NewFileSet()
		file *ast.File
	)
	if file, err = parser.ParseFile(fset, "", src, parseMode); err != nil {
		return
	}
	pkg.Name = "command-line-package"
	appendFile(fset, file, &pkg.Files)
	bast.Packages = append(bast.Packages, pkg)
	return
}

// parseMode is the mode Bast uses for parsing go files.
const parseMode = parser.ParseComments | parser.DeclarationErrors | parser.AllErrors

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

func appendDeclaration(fs *token.FileSet, in ast.Node, out *[]Declaration) {
	switch n := in.(type) {

	case *ast.GenDecl:
		switch n.Tok {
		case token.CONST, token.VAR:

			var (
				lastType  string
			)
			for _, spec := range n.Specs {
				var (
					vs, ok = spec.(*ast.ValueSpec)
					id     *ast.Ident
				)
				if !ok {
					continue
				}
				for i := 0; i < len(vs.Names); i++ {
					var (
						name, typ, val string
						docs, comments []string
					)
					name = vs.Names[i].Name
					appendCommentGroup(vs.Comment, &comments)
					appendCommentGroup(vs.Doc, &docs)
					if vs.Type != nil {
						if id, ok = vs.Type.(*ast.Ident); !ok {
							continue
						}
						typ = id.Name
						lastType = id.Name
					} else if lastType != "" {
						typ = lastType
					}

					if vs.Values != nil {
						switch v := vs.Values[i].(type) {
						case *ast.Ident:
							val = v.Name
						case *ast.BasicLit:
							val, _ = strconv.Unquote(v.Value)
						case *ast.BinaryExpr:
							var (
								lit *ast.BasicLit
							)
							if id, ok = v.X.(*ast.Ident); !ok || id.Name != "iota" {
								continue
							}
							if lit, ok = v.Y.(*ast.BasicLit); !ok {
								continue
							}
							val = fmt.Sprintf("%s %s %s", id.Name, v.Op.String(), lit.Value)
						default:
							continue
						}
					}
					if n.Tok == token.CONST {
						*out = append(*out, &Const{comments, docs, name, typ, val})
					} else if n.Tok == token.VAR {
						*out = append(*out, &Var{comments, docs, name, typ, val})
					}
				}
			}
		case token.TYPE:
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

func appendConst(in *ast.ValueSpec, out *[]Declaration) {
	for i := 0; i < len(in.Names); i++ {
		var (
			val = new(Const)
			id  *ast.Ident
			lit *ast.BasicLit
			ok  bool
		)
		val.Name = in.Names[i].Name
		appendCommentGroup(in.Comment, &val.Comment)
		appendCommentGroup(in.Doc, &val.Doc)

		if id, ok = in.Type.(*ast.Ident); !ok {
			continue
		}
		val.Type = id.Name

		switch v := in.Values[i].(type) {
		case *ast.BasicLit:
			val.Value, _ = strconv.Unquote(v.Value)
		case *ast.BinaryExpr:
			if id, ok = v.X.(*ast.Ident); !ok {
				continue
			}
			if lit, ok = v.Y.(*ast.BasicLit); !ok {
				continue
			}
			val.Value = fmt.Sprintf("%s %s %s", id.Name, v.Op.String(), lit.Value)
		default:
			continue
		}

		*out = append(*out, val)
	}
}

func appendVar(in *ast.ValueSpec, out *[]Declaration) {
	for i := 0; i < len(in.Names); i++ {
		var (
			val = new(Var)
			id  *ast.Ident
			lit *ast.BasicLit
			ok  bool
		)
		val.Name = in.Names[i].Name
		appendCommentGroup(in.Comment, &val.Comment)
		appendCommentGroup(in.Doc, &val.Doc)
		if id, ok = in.Type.(*ast.Ident); !ok {
			continue
		}
		if lit, ok = in.Values[i].(*ast.BasicLit); !ok {
			continue
		}
		val.Type = id.Name
		val.Value, _ = strconv.Unquote(lit.Value)
		*out = append(*out, val)
	}
}

func appendInterface(in *ast.TypeSpec, out *[]Declaration) {
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

func appendStruct(in *ast.TypeSpec, out *[]Declaration) {
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
