package parse

import (
	"fmt"
	"html/template"
	"io"
	"path"
	"strings"

	log "github.com/cihub/seelog"
)

const KitTemplate = `// this file is generated for {{.Name}}
package controllers

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-kit/kit/endpoint"
	{{genimports}}
)

func (*{{service}}) {{.Name}}(request {{parseRequest}}) {{parseResponse}} {
	return nil
}

func MakeEndpoint{{.Name}}(svc {{interface}}) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.({{parseRequest}})
		resp := svc.{{.Name}}(req)
		return resp, nil
	}
}

func DecodeRequest{{.Name}}(ctx context.Context, r *http.Request) (interface{}, error) {
	{{newRequest}}
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return nil, err
	}
	return request, nil
}

func EncodeResponse{{.Name}}(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
`

func (file *File) GenKitFile(Interface *Interface, Func *Func, wr io.Writer) {
	log.Info(fmt.Sprintf("generating %s files ...", Func.Name))
	temp := template.New(fmt.Sprintf("%s.go", Func.Name))
	temp.Funcs(template.FuncMap{
		"package": func() string {
			return path.Base(path.Dir(file.GenPath))
		},
		"service": func() string {
			return file.Package
		},
		"interface": func() string {
			return fmt.Sprintf("%s.%s", path.Base(path.Dir(file.GenPath)), Interface.Name)
		},
		"parseRequest": func() string {
			for index := range Func.Params {
				return Func.Params[index].GoType
			}
			return fmt.Sprintf("%sRequest", Func.Name)
		},
		"parseResponse": func() string {
			for index := range Func.Results {
				return Func.Results[index].GoType
			}
			return fmt.Sprintf("%sResponse", Func.Name)
		},
		"genimports": func() template.HTML {
			imports := strings.Builder{}
			imports.WriteString("\n\t")
			imports.WriteString(fmt.Sprintf(`%s "%s"`, path.Base(path.Dir(file.GenPath)), path.Dir(file.GenPath)))
			for k, v := range file.ImportA {
				imports.WriteString("\n\t")
				imports.WriteString(fmt.Sprintf(`%s "%s"`, k, v))
			}
			return template.HTML(imports.String())
		},
		"newRequest": func() string {
			for index := range Func.Params {
				return fmt.Sprintf(`request := new(%s)`, strings.TrimPrefix(Func.Params[index].GoType, "*"))
			}
			return fmt.Sprintf(`request := new(%s)`, Func.Name)
		},
	})
	kitTemplate, err := temp.Parse(KitTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	kitTemplate.Execute(wr, Func)
}
