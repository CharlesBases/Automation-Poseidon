package template

const implementTemplate = `package {{package}}

import (
	{{imports}}
)
{{range $interfaceIndex, $interface := .Interfaces}}
// this implement is generated for {{.Name}}
var {{$interface | service}}service {{$interface | response}}

type {{$interface | service}} struct{}

func New{{$interface | interfaceName}}() {{$interface | response}} {
	return new({{$interface | service}})
}
{{end}}
// init 
func init() {   {{range $interfaceIndex, $interface := .Interfaces}}
	{{$interface | service}}service = New{{$interface | interfaceName}}()
}
{{end}}`
