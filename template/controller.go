package template

import (
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"path"
	"path/filepath"
	"strings"

	"charlesbases/Automation-Poseidon/utils"
)

var (
	requestParamDepth  = 0
	responseParamDepth = 0
)

var (
	jsonDefaultValue = map[string]interface{}{
		"Number":  0,
		"String":  "",
		"Boolean": false,
		"Object":  nil,
	}
)

func (controller *ControllerInfor) GenerateController(wr io.Writer) {
	utils.Info(fmt.Sprintf("generating %s files ...", controller.Func.Name))
	temp := template.New(fmt.Sprintf("%s.go", controller.Func.Name))
	temp.Funcs(template.FuncMap{
		"package": func() string {
			return path.Base(poseidon.GenInterPath)
		},
		"service": func() string {
			return strings.ToLower(strings.TrimSuffix(controller.InterfaceName, "Service"))
		},
		"interface": func() string {
			return fmt.Sprintf("%s.%s", path.Base(poseidon.PackagePath), controller.InterfaceName)
		},
		"imports":             controller.imports,             // import
		"business":            controller.business,            // 业务层调用
		"jsonDemo":            controller.jsonDemo,            // json 示例
		"requestParse":        controller.requestParse,        // 入参
		"requestDecode":       controller.requestDecode,       // decode request
		"requestCoercive":     controller.requestCoercive,     // request 类型断言
		"responseParse":       controller.responseParse,       // 出参
		"responseAssignment":  controller.responseAssignment,  // response 初始化
		"parseRequestParams":  controller.parseRequestParams,  // request params
		"parseResponseParams": controller.parseResponseParams, // response params
	})
	controllerTemplate, err := temp.Parse(controllerTemplate)
	utils.ThrowCheck(err)
	controllerTemplate.Execute(wr, controller.Func)
}

func (controller *ControllerInfor) imports() template.HTML {
	sb := strings.Builder{}

	sb.WriteString("\n\t" + fmt.Sprintf(`"%s/logics/%s"`, poseidon.ProjectPath, strings.ToLower(controller.Func.Group)))
	sb.WriteString("\n\t" + fmt.Sprintf(`"%s/%s"`, poseidon.ProjectPath, filepath.Base(poseidon.PackagePath)))

	for k, v := range poseidon.Imports {
		sb.WriteString("\n\t")
		if path.Base(v) == k {
			sb.WriteString(fmt.Sprintf(`"%s"`, v))
		} else {
			sb.WriteString(fmt.Sprintf(`%s "%s"`, k, v))
		}
	}
	return template.HTML(sb.String())
}

func (controller *ControllerInfor) requestParse() template.HTML {
	sb := strings.Builder{}
	for i, x := range controller.Func.Params {
		if i != 0 {
			sb.WriteString(", ")
		}
		sb.WriteString(fmt.Sprintf("%s %s", x.Name, x.VariableType))
	}
	return template.HTML(sb.String())
}

func (controller *ControllerInfor) requestDecode() template.HTML {
	sb := strings.Builder{}
	for _, x := range controller.Func.Params {
		if x.Package == "context" || x.Package == "golang.org/x/net/context" {
			continue
		}
		sb.WriteString(fmt.Sprintf(`
			request := new(%s)
			return web.DecodeRequest(r, request)
`,
			strings.TrimPrefix(x.VariableType, "*"),
		))
		break
	}
	return template.HTML(sb.String())
}

func (controller *ControllerInfor) requestCoercive() template.HTML {
	sb := strings.Builder{}
	for i, x := range controller.Func.Params {
		if i != 0 {
			sb.WriteString(", ")
		}
		if x.Package == "context" || x.Package == "golang.org/x/net/context" {
			sb.WriteString(x.Name)
		} else {
			sb.WriteString(fmt.Sprintf("request.(%s)", x.VariableType))
		}
	}
	return template.HTML(sb.String())
}

func (controller *ControllerInfor) responseParse() template.HTML {
	sb := strings.Builder{}
	for _, x := range controller.Func.Results {
		sb.WriteString(fmt.Sprintf("response %s", x.VariableType))
		break
	}
	return template.HTML(sb.String())
}

func (controller *ControllerInfor) responseAssignment() template.HTML {
	sb := strings.Builder{}
	for _, x := range controller.Func.Results {
		sb.WriteString(fmt.Sprintf("response = new(%s)", strings.TrimPrefix(x.VariableType, "*")))
		break
	}
	return template.HTML(sb.String())
}

func (controller *ControllerInfor) business() template.HTML {
	sb := strings.Builder{}
	sb.WriteString(fmt.Sprintf("response = new(%s.%s).%s(%s)",
		strings.ToLower(controller.Func.Group),
		controller.Func.Group,
		controller.Func.Name,
		func() string {
			params := strings.Builder{}
			for i, x := range controller.Func.Params {
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

func (controller *ControllerInfor) parseRequestParams(fields *[]utils.Field) template.HTML {
	requestParamDepth++

	sb := strings.Builder{}
	for _, field := range *fields {
		if requestParamDepth == 1 {
			if field.Package == "context" || field.Package == "golang.org/x/net/context" {
				continue
			}
		}
		for _, structField := range poseidon.Structs[field.Package][field.ProtoType] {
			sb.WriteString(fmt.Sprintf("@apiParam {%s} %s %s\n", structField.JsonType, utils.Snake(structField.Name), structField.Comment))
			if structField.Package != "" {
				sb.WriteString(fmt.Sprintf("%s", controller.parseRequestParams(&[]utils.Field{structField})))
			}
		}
	}
	return template.HTML(sb.String())
}

func (controller *ControllerInfor) parseResponseParams(fields *[]utils.Field) template.HTML {
	responseParamDepth++

	sb := strings.Builder{}
	for _, field := range *fields {
		for _, structField := range poseidon.Structs[field.Package][field.ProtoType] {
			if responseParamDepth == 1 {
				if structField.Name != "Results" {
					continue
				}
			}
			sb.WriteString(fmt.Sprintf("@apiSuccess {%s} %s %s\n", structField.JsonType, utils.Snake(structField.Name), structField.Comment))
			if structField.Package != "" {
				sb.WriteString(fmt.Sprintf("%s", controller.parseResponseParams(&[]utils.Field{structField})))
			}
		}
	}
	return template.HTML(sb.String())
}

func (controller *ControllerInfor) jsonDemo(fields *[]utils.Field) template.HTML {
	jsonDatas := controller.decode(fields)
	datas, _ := json.MarshalIndent(jsonDatas, "", "\t")
	return template.HTML(string(datas))
}

func (controller *ControllerInfor) decode(fields *[]utils.Field) interface{} {
	jsonDatas := make(map[string]interface{}, 0)
	for _, param := range *fields {
		if param.Package == "context" || param.Package == "golang.org/x/net/context" {
			continue
		}
		for _, field := range poseidon.Structs[param.Package][param.ProtoType] {
			jsonDatas[utils.Snake(field.Name)] = controller.parseField(&field)
		}
		break
	}
	return jsonDatas
}

func (controller *ControllerInfor) parseField(field *utils.Field) interface{} {
	switch strings.HasPrefix(field.ProtoType, "repeated ") {
	case false:
		if field.Package != "" {
			return controller.decode(&[]utils.Field{*field})
		}
		return jsonDefaultValue[field.JsonType]
	case true:
		field.ProtoType = strings.TrimPrefix(field.ProtoType, "repeated ")
		if field.Package != "" {
			data := make([]interface{}, 1)
			data[0] = controller.decode(&[]utils.Field{*field})
			return data
		} else {
			data := make([]interface{}, 1)
			data[0] = jsonDefaultValue[strings.TrimPrefix(field.JsonType, "[]")]
			return data
		}
	}
	return nil
}
