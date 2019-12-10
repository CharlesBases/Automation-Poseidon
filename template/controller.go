package template

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"path"
	"path/filepath"
	"strings"

	log "github.com/cihub/seelog"

	"charlesbases/Automation-Poseidon/utils"
)

var (
	jsonDefaultValue = map[string]interface{}{
		"Number":  0,
		"String":  "",
		"Boolean": false,
		"Object":  nil,
	}
)

func (infor *Infor) GenerateController(wr io.Writer) {
	log.Info(fmt.Sprintf("generating %s files ...", infor.Func.Name))
	temp := template.New(fmt.Sprintf("%s.go", infor.Func.Name))
	temp.Funcs(template.FuncMap{
		"package": func() string {
			return path.Base(infor.File.GenInterPath)
		},
		"service": func() string {
			return strings.ToLower(strings.TrimSuffix(infor.Interface.Name, "Service"))
		},
		"interface": func() string {
			return fmt.Sprintf("%s.%s", path.Base(infor.File.PackagePath), infor.Interface.Name)
		},
		"imports":             infor.imports,             // import
		"business":            infor.business,            // 业务层调用
		"jsonDemo":            infor.jsonDemo,            // json 示例
		"requestParse":        infor.requestParse,        // 入参
		"requestDecode":       infor.requestDecode,       // decode request
		"requestCoercive":     infor.requestCoercive,     // request 类型断言
		"responseParse":       infor.responseParse,       // 出参
		"responseAssignment":  infor.responseAssignment,  // response 初始化
		"parseRequestParams":  infor.parseRequestParams,  // request params
		"parseResponseParams": infor.parseResponseParams, // response params
	})
	controllerTemplate, err := temp.Parse(controllerTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	controllerTemplate.Execute(wr, infor.Func)
}

func (infor *Infor) imports() template.HTML {
	sb := strings.Builder{}

	sb.WriteString("\n\t" + fmt.Sprintf(`"%s/logics/%s"`, infor.File.ProjectPath, strings.ToLower(infor.Func.Group)))
	sb.WriteString("\n\t" + fmt.Sprintf(`"%s/%s"`, infor.File.ProjectPath, filepath.Base(infor.File.PackagePath)))

	for k, v := range infor.File.ImportA {
		sb.WriteString("\n\t")
		if path.Base(v) == k {
			sb.WriteString(fmt.Sprintf(`"%s"`, v))
		} else {
			sb.WriteString(fmt.Sprintf(`%s "%s"`, k, v))
		}
	}
	return template.HTML(sb.String())
}

func (infor *Infor) requestParse() template.HTML {
	sb := strings.Builder{}
	for i, x := range infor.Func.Params {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%s %s", x.Name, x.GoType))
	}
	return template.HTML(sb.String())
}

func (infor *Infor) requestDecode() template.HTML {
	sb := strings.Builder{}
	for _, x := range infor.Func.Params {
		if x.Package == "context" || x.Package == "golang.org/x/net/context" {
			continue
		}
		sb.WriteString(fmt.Sprintf(`
			request := new(%s)
			return web.DecodeRequest(r, request)
`,
			strings.TrimPrefix(x.GoType, "*"),
		))
		break
	}
	return template.HTML(sb.String())
}

func (infor *Infor) requestCoercive() template.HTML {
	sb := strings.Builder{}
	for i, x := range infor.Func.Params {
		if i != 0 {
			sb.WriteString(", ")
		}
		if x.Package == "context" || x.Package == "golang.org/x/net/context" {
			sb.WriteString(x.Name)
		} else {
			sb.WriteString(fmt.Sprintf("request.(%s)", x.GoType))
		}
	}
	return template.HTML(sb.String())
}

func (infor *Infor) responseParse() template.HTML {
	sb := strings.Builder{}
	for _, x := range infor.Func.Results {
		sb.WriteString(fmt.Sprintf("response %s", x.GoType))
		break
	}
	return template.HTML(sb.String())
}

func (infor *Infor) responseAssignment() template.HTML {
	sb := strings.Builder{}
	for _, x := range infor.Func.Results {
		sb.WriteString(fmt.Sprintf("response = new(%s)", strings.TrimPrefix(x.GoType, "*")))
		break
	}
	return template.HTML(sb.String())
}

func (infor *Infor) business() template.HTML {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("response = new(%s.%s).%s(%s)",
		strings.ToLower(infor.Func.Group),
		infor.Func.Group,
		infor.Func.Name,
		func() string {
			params := strings.Builder{}
			for i, x := range infor.Func.Params {
				if i != 0 {
					params.WriteString(", ")
				}
				params.WriteString(x.Name)
			}
			return params.String()
		}(),
	))
	return template.HTML(sb.String())
}

func (infor *Infor) parseRequestParams(fields []utils.Field) template.HTML {
	sb := strings.Builder{}
	for _, requestParam := range fields {
		if requestParam.Package == "context" || requestParam.Package == "golang.org/x/net/context" {
			continue
		}
		for _, field := range infor.File.Structs[requestParam.Package][requestParam.ProtoType] {
			sb.WriteString(fmt.Sprintf("@apiParam {%s} %s %s\n", field.JsonType, utils.Snake(field.Name), field.Comment))
			if field.Package != "" {
				sb.WriteString(fmt.Sprintf("%s", infor.parseRequestParams([]utils.Field{field})))
			}
		}
		break
	}
	return template.HTML(sb.String())
}

func (infor *Infor) parseResponseParams(fields []utils.Field) template.HTML {
	sb := strings.Builder{}
	for _, responseParam := range fields {
		for _, responseField := range infor.File.Structs[responseParam.Package][responseParam.ProtoType] {
			if responseField.Name == "Results" && responseField.Package != "" {
				for _, resultsField := range infor.File.Structs[responseField.Package][responseField.ProtoType] {
					sb.WriteString(fmt.Sprintf("@apiSuccess {%s} %s %s\n", resultsField.JsonType, utils.Snake(resultsField.Name), resultsField.Comment))
					if resultsField.Package != "" {
						sb.WriteString(fmt.Sprintf("%s", infor.parseResponseParams([]utils.Field{resultsField})))
					}
				}
			}
		}
		break
	}
	return template.HTML(sb.String())
}

func (infor *Infor) jsonDemo(fields []utils.Field) template.HTML {
	jsonDatas := infor.decode(fields)
	datas, _ := json.MarshalIndent(jsonDatas, "", "\t")
	return template.HTML(string(datas))
}

func (infor *Infor) decode(fields []utils.Field) interface{} {
	jsonDatas := make(map[string]interface{}, 0)
	for _, param := range fields {
		if param.Package == "context" || param.Package == "golang.org/x/net/context" {
			continue
		}
		for _, field := range infor.File.Structs[param.Package][param.ProtoType] {
			jsonDatas[utils.Snake(field.Name)] = infor.parseField(field)
		}
		break
	}
	return jsonDatas
}

func (infor *Infor) parseField(field utils.Field) interface{} {
	switch strings.HasPrefix(field.ProtoType, "repeated ") {
	case false:
		if field.Package != "" {
			return infor.decode([]utils.Field{field})
		}
		return jsonDefaultValue[field.JsonType]
	case true:
		field.ProtoType = strings.TrimPrefix(field.ProtoType, "repeated ")
		if field.Package != "" {
			data := make([]interface{}, 1)
			data[0] = infor.decode([]utils.Field{field})
			return data
		} else {
			data := make([]interface{}, 1)
			data[0] = jsonDefaultValue[strings.TrimPrefix(field.JsonType, "[]")]
			return data
		}
	}
	return nil
}
