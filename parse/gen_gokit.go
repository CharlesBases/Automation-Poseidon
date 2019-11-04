package parse

import (
	"fmt"
	"html/template"
	"io"

	log "github.com/cihub/seelog"
)

const KitTemplate = `// this file is generated for {{.Name}}
func (*{{service}}) {{.Name}}(request *{{parseRequest}}) *{{parseRequest}} {
	return nil
}

func MakeEndpoint{{.Name}}(svc {{interface}}) endpoint.Endpoint {
	return func(ctx context.Context, request interface{}) (interface{}, error) {
		req := request.(*{{parseRequest}})
		resp := svc.{{.Name}}(req)
		return resp, nil
	}
}

func DecodeRequest{{.Name}}(ctx context.Context, r *http.Request) (interface{}, error) {
	request := new({{parseRequest}})
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		return nil, err
	}
	return request, nil
}

func EncodeResponseAdds(ctx context.Context, w http.ResponseWriter, response interface{}) error {
	return json.NewEncoder(w).Encode(response)
}
`

func (file *File) GenKitFile(Interface *Interface, Func *Func, wr io.Writer) {
	log.Info(fmt.Sprintf("generating %s files ...", Func.Name))
	temp := template.New(fmt.Sprintf("%s.go", Func.Name))
	temp.Funcs(template.FuncMap{
		"service": func() string {
			return file.Package
		},
		"interface": func() string {
			return Interface.Name
		},
		"parseRequest": func() string {
			for index := range Func.Params {
				return Func.Params[index].ProtoType
			}
			return fmt.Sprintf("%sRequest", Func.Name)
		},
		"parseResponse": func() string {
			for index := range Func.Results {
				return Func.Results[index].ProtoType
			}
			return fmt.Sprintf("%sResponse", Func.Name)
		},
	})
	kitTemplate, err := temp.Parse(KitTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	kitTemplate.Execute(wr, Func)
}
