package parse

import (
	"fmt"
	"html/template"
	"io"
	"path"
	"path/filepath"
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

/*
@api {post} url title
@apiVersion 1.0.0
@apiGroup {{parsegroup}}
@apiDescription 描述

@apiParam {Object} p 请求参数
{{parseRequestParams .Params}}
@apiParamExample {json} Request-Example:
{

}

@apiSuccess {Object} results 返回结果
{{parseResponseParams .Results}}
@apiSuccessExample {json} Response-Example:
{

}
*/
func (*{{service}}) {{.Name}}({{requestParse}}) ({{responseParse}}) {
	// response.Results
	{{results}}
	// response.Error
	Error := new(web.Error)
	defer func() {
		if Error.ErrNo != 0 {
			response = {{responseAssignment}}{
				Error: core.Error{
					ErrNo:  Error.ErrNo,
					ErrMsg: Error.ErrMsg,
				},
			}
		} else {
			response = {{responseAssignment}}{
				Results: results,
			}
		}
	}()
	// validator
	if err := validator.New().Struct(request); err != nil {
		log.Error("json validator error: ", err)
		Error.WebError(web.ParamsErr)
		return
	}

	// Do something

	return 
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
		resp := svc.{{.Name}}({{requestCoercive}})
		return resp, nil
	}
}

func DecodeRequest{{.Name}}(ctx context.Context, r *http.Request) (interface{}, error) {
	{{requestDecode}}
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
			return path.Base(path.Dir(file.GenProtoPath))
		},
		"service": func() string {
			return strings.ToLower(strings.ReplaceAll(Interface.Name, "Service", ""))
		},
		"genimports": func() template.HTML {
			imports := strings.Builder{}
			imports.WriteString("\n\t")
			imports.WriteString(fmt.Sprintf(`"%s"`, strings.ReplaceAll(func() string {
				if file.ProjectPath != "" {
					return fmt.Sprintf("%s/%s", file.ProjectPath, filepath.Base(file.PackagePath))
				} else {
					return file.PackagePath
				}
			}(), `\`, `/`)))
			for k, v := range file.ImportA {
				imports.WriteString("\n\t")
				imports.WriteString(fmt.Sprintf(`%s "%s"`, k, v))
			}
			return template.HTML(imports.String())
		},
		"interface": func() string {
			return fmt.Sprintf("%s.%s", filepath.Base(file.PackagePath), Interface.Name)
		},
		"requestParse": func() string {
			request := strings.Builder{}
			for _, x := range Func.Params {
				request.WriteString(fmt.Sprintf("request %s", x.GoType))
				break
			}
			return request.String()
		},
		"requestDecode": func() string {
			request := strings.Builder{}
			if len(Func.Params) != 0 {
				request.WriteString(fmt.Sprintf(`request := new(%s)
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		log.Error("decode request error: ", err)
	}
	return request, nil`,
					func() string {
						requestParam := strings.Builder{}
						for _, param := range Func.Params {
							requestParam.WriteString(strings.TrimPrefix(param.GoType, "*"))
						}
						return requestParam.String()
					}(),
				))
			} else {
				request.WriteString("return nil, nil")
			}
			return request.String()
		},
		"requestCoercive": func() string {
			request := strings.Builder{}
			for _, x := range Func.Params {
				request.WriteString(fmt.Sprintf("request.(%s)", x.GoType))
				break
			}
			return request.String()
		},
		"responseParse": func() string {
			response := strings.Builder{}
			for _, x := range Func.Results {
				response.WriteString(fmt.Sprintf("response %s", x.GoType))
				break
			}
			return response.String()
		},
		"responseAssignment": func() template.HTML {
			response := strings.Builder{}
			for _, param := range Func.Results {
				response.WriteString(strings.ReplaceAll(param.GoType, "*", "&"))
				break
			}
			return template.HTML(response.String())
		},
		"results": func() string {
			for _, response := range Func.Results {
				for _, Struct := range file.Structs {
					if response.ProtoType == Struct.Name {
						for _, field := range Struct.Fields {
							if strings.HasPrefix(field.GoType, "*") {
								return fmt.Sprintf("results := new(%s.%s)", filepath.Base(field.Package), strings.TrimPrefix(field.GoType, "*"))
							}
						}
					}
				}
				break
			}
			return "var results interface{}"
		},
		"parsegroup": func() string {
			group := strings.Builder{}
			for i, x := range []rune(Func.Name) {
				if i != 0 && x < 91 && x > 64 {
					break
				}
				group.WriteString(string(x))
			}
			return group.String()
		},
		"parseRequestParams":  file.parseRequestParams,
		"parseResponseParams": file.parseResponseParams,
	})
	kitTemplate, err := temp.Parse(KitTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	kitTemplate.Execute(wr, Func)
}

func (file *File) parseRequestParams(fields []Field) string {
	doc := strings.Builder{}

	RequestStruct := new(Struct)

	for _, requestParam := range fields {
		for _, Struct := range file.Structs {
			if requestParam.ProtoType == Struct.Name {
				RequestStruct = &Struct
				break
			}
		}
		break
	}

	for _, field := range RequestStruct.Fields {
		doc.WriteString(fmt.Sprintf("@apiParam {%s} %s %s\n", field.JsonType, ensnake(field.Name), field.Comment))
		if field.Package != "" {
			doc.WriteString(fmt.Sprintf("%s", file.parseRequestParams([]Field{field})))
		}
	}

	return doc.String()
}

func (file *File) parseResponseParams(fields []Field) string {
	doc := strings.Builder{}

	ResponseStruct := new(Struct)

	for _, responseParam := range fields {
		for _, Struct := range file.Structs {
			if strings.TrimPrefix(responseParam.ProtoType, "repeated ") == Struct.Name {
				ResponseStruct = &Struct
				break
			}
		}
		break
	}

	for _, field := range ResponseStruct.Fields {
		if field.Name == "Results" && field.Package != "" {
			for _, Struct := range file.Structs {
				if field.ProtoType == Struct.Name {
					ResponseStruct = &Struct
					break
				}
			}
		}
	}

	for _, field := range ResponseStruct.Fields {
		doc.WriteString(fmt.Sprintf("@apiSuccess {%s} %s %s\n", field.JsonType, ensnake(field.Name), field.Comment))
		if field.Package != "" {
			doc.WriteString(fmt.Sprintf("%s", file.parseResponseParams([]Field{field})))
		}
	}

	return doc.String()
}

// AaaBbb to aaa_bbb
func ensnake(source string) string {
	builder := strings.Builder{}
	ascll := []rune(source)
	for key, word := range ascll {
		if word <= 90 {
			if key != 0 {
				if word != 68 || ascll[key-1] != 73 {
					builder.WriteString("_")
				}
			}
			builder.WriteString(strings.ToLower(string(word)))
		} else {
			builder.WriteString(string(word))
		}
	}
	return builder.String()
}
