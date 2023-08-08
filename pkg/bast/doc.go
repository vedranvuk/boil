// Copyright 2023 Vedran Vuk. All rights reserved.
// Use of this source code is governed by a MIT
// license that can be found in the LICENSE file.

// Package bast implements a (B)astard|(B)oilerplated AST; an object model of a 
// stripped down go/ast parse hierarchy used for easier analysis of go source 
// files from templating engines like text/template.
//
// Currently it reads package and file information and top level declarations
// of which following is supported:
//   * Interfaces and their method sets.
//   * Structs and their fields and method sets. (WIP)
//   * Const and var declarations.
//
// Bast makes no use of type checking; it is not a compiler, it just extracts 
// text tokens.
package bast