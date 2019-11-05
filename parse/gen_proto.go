package parse

import (
	"html/template"
	"io"

	log "github.com/cihub/seelog"
)

const ProtoTemplate = `// this file is generated from {{.PkgPath}}
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

func (file *File) GenProtoFile(wr io.Writer) {
	log.Info("generating .proto files ...")
	temp := template.New("pb.proto")
	temp.Funcs(template.FuncMap{
		"index": func(i int) int {
			return i + 1
		},
		"unescaped": func(x string) template.HTML {
			return template.HTML(x)
		},
	})
	protoTemplate, err := temp.Parse(ProtoTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	protoTemplate.Execute(wr, file)
}
