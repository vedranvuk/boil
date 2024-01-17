module github.com/vedranvuk/boil

go 1.21.0

require (
	github.com/adrg/xdg v0.4.0
	github.com/vedranvuk/bast v0.0.0-00010101000000-000000000000
	github.com/vedranvuk/cmdline v0.0.0-20230731121628-0e879a0d21b4
	github.com/vedranvuk/tmpl v0.0.0-00010101000000-000000000000
)

require (
	github.com/stretchr/testify v1.8.4 // indirect
	golang.org/x/mod v0.12.0 // indirect
	golang.org/x/sys v0.11.0 // indirect
	golang.org/x/tools v0.12.0 // indirect
)

replace github.com/vedranvuk/cmdline => ../cmdline

replace github.com/vedranvuk/bast => ../bast

replace github.com/vedranvuk/tmpl => ../tmpl
