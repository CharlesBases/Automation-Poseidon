package template

const protoTemplate = `// this file is generated from {{.PkgPath}}
syntax = "proto3";

package {{.Package}};
{{range $interfaceIndex, $interface := .Interfaces}}
service {{.Name}} {
{{range $funcsIndex, $func := .Funcs}}    rpc {{.Name}} ({{.Name}}Req_) returns ({{.Name}}Resp_) {} 
{{end}}}
{{range $funcsIndex, $func := .Funcs}}
message {{.Name}}Req_ {
{{range $paramsIndex, $param := .Params}}    {{.ProtoType | unescaped}} {{.Name}} = {{$paramsIndex | index}};
{{end}}}

message {{$func.Name}}Resp_ {
{{range $resultsIndex, $vresult:= .Results}}    {{.ProtoType | unescaped}} {{.Name}} = {{$resultsIndex | index}};
{{end}}}
{{end}}{{end}}{{range $structsIndex, $struct := .Structs}}
message {{$struct.Name}} {
{{range $fieldsIndex, $field := .Fields}}    {{.ProtoType | unescaped}} {{.Name}} = {{$fieldsIndex | index}};
{{end}}}
{{end}}
`
