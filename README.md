# Boil

**experimental**

Boil is a boilerplating/templating tool for Go that creates complete project 
boilerplates or smaller fragments of a codebase.

___

It takes snapshots of some source directory and optionaly all its child 
directories and files retaining file system hierarchy and packages it into a 
template.

Template file names and content can be parametrized. Content can be templated 
using `text/template`, input data comes from stdin prompts, command line 
arguments, AST of some go file or package, other input files, etc.

Templates are stored into repositories by default and boil maintains a default
repository. Templates can be addressed by a path relative to the loaded 
repository or an absolute path to a template. Custom repositories can be loaded.
Both templates and repositories can separately be versioned as they are just
a directory on a file system. Loading of repositories or templates from network
might be implemented later.

Templates can have sub-templates and define groups of them so combinations of 
parent and child templates can be executed as a single template, enabling 
template modularity.

Custom commands can be defined at various stages of template execution so data
can be generated externally and optionally cleaned up after.

A metafile named `boil.json` defines a template and resides in the root of a 
template structure.

Up to date help is in the tool itself and reachable via `boil help`.

## Installation

To install boil type `go install github.com/vedranvuk/boil`.

## License

MIT
