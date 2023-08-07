package {{.Vars.PackageName}}

var _{{.Vars.TypeName}}_strings = []string{
	{{ range $const := ConstsOfType "" .Vars.TypeName }}"{{ $const.Name }}",
	{{ end -}}
}

func (self {{.Vars.TypeName}}) String() string { return _{{.Vars.TypeName}}_strings[int(self)] }
