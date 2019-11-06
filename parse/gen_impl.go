package parse

import (
	"fmt"
	"html/template"
	"io"
	"path"
	"strings"

	log "github.com/cihub/seelog"
)

const ImplTemplate = `package controllers

import (
	{{genimports}}
)
{{range $interfaceIndex, $interface := .Interfaces}}
// this implement is generated for {{.Name}}
type {{$interface | service}} struct{}

func New{{$interface | interfaceName}}() {{$interface | response}} {
	return new({{$interface | service}})
}
{{end}}
`

func (file *File) GenImplFile(Package *Package, wr io.Writer) {
	log.Info("generating implement files ...")
	temp := template.New("implement.go")
	temp.Funcs(template.FuncMap{
		"genimports": func() template.HTML {
			imports := strings.Builder{}
			imports.WriteString(fmt.Sprintf(`%s "%s"`, path.Base(path.Dir(file.GenPath)), path.Dir(file.GenPath)))
			return template.HTML(imports.String())
		},
		"service": func(Interface Interface) string {
			return strings.ToLower(strings.Replace(Interface.Name, "Service", "", 1))
		},
		"interfaceName": func(Interface Interface) string {
			return Interface.Name
		},
		"response": func(Interface Interface) string {
			return fmt.Sprintf("%s.%s", path.Base(path.Dir(file.GenPath)), Interface.Name)
		},
	})
	implTemplate, err := temp.Parse(ImplTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	implTemplate.Execute(wr, file)
}
