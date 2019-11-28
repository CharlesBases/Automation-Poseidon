package template

const protoTemplate = `// this file is generated from {{.PackagePath}}
syntax = "proto3";

package {{.ProtoPackage}};
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
{{end}}{{end}}{{range $structsPackage, $structs := .Structs}} {{range $structName, $fields := $structs}}
message {{$structName}} {
{{range $fieldsIndex, $field := $fields}}    {{.ProtoType | unescaped}} {{.Name}} = {{$fieldsIndex | index}};
{{end}}}
{{end}}
{{end}}
`
