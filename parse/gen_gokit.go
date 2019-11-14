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

	"gitlab.ifchange.com/bot/gokitcommon/web"
	"gopkg.in/go-playground/validator.v9"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	{{genimports}}
)

func (*{{service}}) {{.Name}}(request {{parseRequest}}) ({{parseResponse}}) {
	defer func() {
		if response.Error.ErrNo != 0 {
			response.Results = nil
		}
	}()
	if err := validator.New().Struct(request); err != nil {
		response.Error.WebError(web.ParamsErr)
		return response
	}

	return response
}

func {{.Name}}() http.Handler {
	return httptransport.NewServer(
		MakeEndpoint{{.Name}}({{service}}service),
		DecodeRequest{{.Name}},
		EncodeResponse{{.Name}},
	)
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
	return web.Encode(w, response)
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
				return fmt.Sprintf("response %s", Func.Results[index].GoType )
			}
			return fmt.Sprintf("response %sResponse", Func.Name)
		},
		"newResponse": func() string {
			for index := range Func.Results {
				return fmt.Sprintf("new(%s)", strings.TrimLeft(Func.Results[index].GoType, "*"))
			}
			return fmt.Sprintf("new(%sResponse)", Func.Name)
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
