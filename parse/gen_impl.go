package parse

import (
	"fmt"
	"html/template"
	"io"
	"path/filepath"
	"strings"

	log "github.com/cihub/seelog"
)

const ImplTemplate = `package {{package}}

import (
	{{genimports}}
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
{{end}}
`

func (file *File) GenImplFile(wr io.Writer) {
	log.Info("generating implement files ...")
	temp := template.New("implement.go")
	temp.Funcs(template.FuncMap{
		"package": func() string {
			return filepath.Base(file.GenInterPath)
		},
		"service": func(Interface Interface) string {
			return strings.ToLower(strings.TrimRight(Interface.Name, "Service"))
		},
		"response": func(Interface Interface) string {
			return fmt.Sprintf("%s.%s", filepath.Base(file.PackagePath), Interface.Name)
		},
		"genimports": func() template.HTML {
			imports := strings.Builder{}
			imports.WriteString(fmt.Sprintf(`"%s"`, strings.ReplaceAll(func() string {
				if file.ProjectPath != "" {
					return fmt.Sprintf("%s/%s", file.ProjectPath, filepath.Base(file.PackagePath))
				} else {
					return file.PackagePath
				}
			}(), `\`, `/`)))
			return template.HTML(imports.String())
		},
		"interfaceName": func(Interface Interface) string {
			return Interface.Name
		},
	})
	implTemplate, err := temp.Parse(ImplTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	implTemplate.Execute(wr, file)
}
