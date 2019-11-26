package parse

import (
	"encoding/json"
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
	"encoding/json"
	"net/http"

	"gitlab.ifchange.com/bot/gokitcommon/log"
	"gitlab.ifchange.com/bot/gokitcommon/web"
	"gopkg.in/go-playground/validator.v9"

	"github.com/go-kit/kit/endpoint"
	httptransport "github.com/go-kit/kit/transport/http"
	{{genimports}}
)

/**
@api {post} url title
@apiVersion 1.0.0
@apiGroup {{.Group}}
@apiDescription 描述

@apiParam {Object} p 请求参数
{{parseRequestParams .Params}}
@apiParamExample {json} Request-Example:
{{decodeJson .Params}}

@apiSuccess {Object} results 返回结果
{{parseResponseParams .Results}}
@apiSuccessExample {json} Response-Example:
{{decodeJson .Results}}

*/
func (*{{service}}) {{.Name}}({{requestParse}}) ({{responseParse}}) {
	defer func() {
		if response.ErrNo != 0 {
			response.Results = nil
		}
	}()
	// response
	{{responseAssignment}}
	// validator
	if err := validator.New().Struct(request); err != nil {
		log.Error("json validator error: ", err)
		response.ErrNo, response.ErrMsg = web.WebError(web.ParamsErr)
		return
	}
	// Do something
	{{business}}

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
	return web.EncodeResponse(w, response)
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
			imports.WriteString(fmt.Sprintf(`"%s/logics/%s"`, file.ProjectPath, strings.ToLower(Func.Group)))

			imports.WriteString("\n\n\t")
			imports.WriteString(fmt.Sprintf(`"%s/%s"`, file.ProjectPath, filepath.Base(file.PackagePath)))

			for k, v := range file.ImportA {
				imports.WriteString("\n\t")
				if path.Base(v) == k {
					imports.WriteString(fmt.Sprintf(`"%s"`, v))
				} else {
					imports.WriteString(fmt.Sprintf(`%s "%s"`, k, v))
				}
			}
			return template.HTML(imports.String())
		},
		"interface": func() string {
			return fmt.Sprintf("%s.%s", filepath.Base(file.PackagePath), Interface.Name)
		},
		"requestParse": func() string {
			request := strings.Builder{}
			for i, x := range Func.Params {
				if i != 0 {
					request.WriteString(", ")
				}
				request.WriteString(fmt.Sprintf("%s %s", x.Name, x.GoType))
			}
			return request.String()
		},
		"requestDecode": func() template.HTML {
			request := strings.Builder{}
			if len(Func.Params) != 0 {
				request.WriteString(fmt.Sprintf(`request := new(%s)
	if err := json.NewDecoder(r.Body).Decode(request); err != nil {
		log.Error("decode request error: ", err)
	}
	return request, nil`,
					func() string {
						requestParam := strings.Builder{}
						for _, x := range Func.Params {
							if x.Package == "context" || x.Package == "golang.org/x/net/context" {
								continue
							}
							requestParam.WriteString(strings.TrimPrefix(x.GoType, "*"))
						}
						return requestParam.String()
					}(),
				))
			} else {
				request.WriteString("return nil, nil")
			}
			return template.HTML(request.String())
		},
		"requestCoercive": func() string {
			request := strings.Builder{}
			for i, x := range Func.Params {
				if i != 0 {
					request.WriteString(", ")
				}
				if x.Package == "context" || x.Package == "golang.org/x/net/context" {
					request.WriteString(x.Name)
				} else {
					request.WriteString(fmt.Sprintf("request.(%s)", x.GoType))
				}
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
				if strings.HasPrefix(param.GoType, "*") {
					response.WriteString(fmt.Sprintf("response = new(%s)", strings.TrimPrefix(param.GoType, "*")))
					break
				} else {
					response.WriteString(fmt.Sprintf("response = %s", param.GoType))
					break
				}
			}
			return template.HTML(response.String())
		},
		"business": func() template.HTML {
			business := strings.Builder{}
			business.WriteString(fmt.Sprintf("response = new(%s.%s).%s(%s)",
				strings.ToLower(Func.Group),
				Func.Group,
				Func.Name,
				func() string {
					params := strings.Builder{}
					for i, x := range Func.Params {
						if i != 0 {
							params.WriteString(", ")
						}
						params.WriteString(x.Name)
					}
					return params.String()
				}(),
			))
			return template.HTML(business.String())
		},
		"decodeJson":          file.decodeJson,
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

func (file *File) parseRequestParams(fields []Field) template.HTML {
	doc := strings.Builder{}

	for _, requestParam := range fields {
		for _, field := range file.Structs[requestParam.Package][requestParam.ProtoType] {
			doc.WriteString(fmt.Sprintf("@apiParam {%s} %s %s\n", field.JsonType, Snake(field.Name), field.Comment))
			if field.Package != "" {
				doc.WriteString(fmt.Sprintf("%s", file.parseRequestParams([]Field{field})))
			}
		}
		break
	}

	return template.HTML(doc.String())
}

func (file *File) parseResponseParams(fields []Field) string {
	doc := strings.Builder{}

	for _, responseParam := range fields {
		for _, responseField := range file.Structs[responseParam.Package][responseParam.ProtoType] {
			if responseField.Name == "Results" && responseField.Package != "" {
				for _, resultsField := range file.Structs[responseField.Package][responseField.ProtoType] {
					doc.WriteString(fmt.Sprintf("@apiSuccess {%s} %s %s\n", resultsField.JsonType, Snake(resultsField.Name), resultsField.Comment))
					if resultsField.Package != "" {
						doc.WriteString(fmt.Sprintf("%s", file.parseResponseParams([]Field{resultsField})))
					}
				}
			}
		}
		break
	}

	return doc.String()
}

func (file *File) decodeJson(fields []Field) template.HTML {
	jsonDatas := file.jsonDemo(fields)
	datas, _ := json.MarshalIndent(jsonDatas, "", "\t")
	return template.HTML(string(datas))
}

func (file *File) jsonDemo(fields []Field) interface{} {
	jsonData := make(map[string]interface{}, 0)

	for _, param := range fields {
		for _, field := range file.Structs[param.Package][param.ProtoType] {
			jsonData[Snake(field.Name)] = file.parseFieldValue(field)
		}
		break
	}

	return jsonData
}

func (file *File) parseFieldValue(field Field) interface{} {
	switch strings.HasPrefix(field.ProtoType, "repeated ") {
	case false:
		if field.Package != "" {
			return file.jsonDemo([]Field{field})
		}
		return jsonDefaultValue[field.JsonType]
	case true:
		field.ProtoType = strings.ReplaceAll(field.ProtoType, "repeated ", "")

		if field.Package != "" {
			data := make([]interface{}, 1)
			data[0] = file.jsonDemo([]Field{field})
			return data
		} else {
			data := make([]interface{}, 1)
			data[0] = jsonDefaultValue[strings.ReplaceAll(field.JsonType, "[]", "")]
			return data
		}

	}
	return nil
}

// AaaBbb to aaa_bbb
func Snake(source string) string {
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
