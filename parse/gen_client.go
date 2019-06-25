package parse

import (
	"fmt"
	"io"
	"strings"
	"text/template"

	log "github.com/cihub/seelog"
)

const ServiceClientTemplate = `// this file is generated from {{.PkgPath}} {{$pkg := .PkgPath | packageSort}} {{$Package := .Package}}
package {{.Package}}
import (
	"fmt"
	{{range $importIndex, $import := .ImportA}}{{generateImport $importIndex $import}}
	{{end}}

	proto "github.com/CharlesBases/proto/parse"

	log "github.com/cihub/seelog"
	_struct "github.com/golang/protobuf/ptypes/struct"
	"github.com/micro/go-micro/client"
)
{{range $interfaceIndex, $interface := .Interfaces}}
func New{{.Name}}Client_({{.Name}} {{.Name}}, opts ...client.CallOption) {{$pkg}}.{{.Name}} {
	return &{{.Name}}ClientImpl{
		{{.Name}}:  {{.Name}},
		opts:       opts,
	}
}

type {{.Name}}ClientImpl struct {
	{{.Name}}   {{.Name}}
	opts        []client.CallOption
}
{{range $funcsIndex, $func := .Funcs}} {{$ParamsLen := .Params | len | funcReduce}} {{$ResultsLen := .Results | len | funcReduce}}
func ({{$interface.Name}} *{{$interface.Name}}ClientImpl) {{.Name}}(ctx context.Context, {{range $paramsIndex, $param := .Params}}{{.Name}} {{.GoType}} {{if ne $paramsIndex  $ParamsLen }},{{end}} {{end}}) ({{range $resultsIndex, $result := .Results}} {{.Name}} {{.GoType}}{{if ne $resultsIndex $ResultsLen }},{{end}} {{end}}) {
	serviceRequest := &{{.Name}}Req_{ {{range $paramsIndex, $param := .Params}}
		{{convertClientRequest . ""}}{{end}}
	}

	serviceResponse, serviceError := {{$interface.Name}}.{{$interface.Name}}.{{.Name}}(ctx, serviceRequest, {{$interface.Name}}.opts...)
	if serviceError != nil {
		log.Error(serviceError)
		panic(serviceError.Error())
	}
	{{range $resultsIndex, $result := .Results}}
	{{.Name}} = {{convertClientResponse . "serviceResponse"}}
	{{end}}
    {{if eq $ResultsLen -1}} = serviceResponse{{end}}
	return {{range $resultsIndex, $result := .Results}} {{.Name}}{{if ne $resultsIndex $ResultsLen }},{{end}} {{end}}
} 
{{end}}
{{end}}
{{range $structsIndex, $struct := .Structs}}
{{if $struct.IsRecursion }}
func clientModel{{.Name}}(value  *{{.Name}}) model.{{.Name}} {
	result := model.{{.Name}}{} 
	{{range $fieldsIndex, $field := .Fields}}
	result.{{.Name}} = {{convertClientRequest . "value"}}
	{{end}}
	return result
}
{{end}}
{{end}}
`

func (file *File) GenClient(wr io.Writer) {
	log.Info("generating client file ...")
	t := template.New("pb.client.go")
	t.Funcs(template.FuncMap{
		"funcReduce": func(i int) int {
			return i - 1
		},
		"service": func(n string) string {
			if strings.HasSuffix(n, "Service") {
				return n
			}
			return n + "Service"
		},
		"packageSort":           packageSort,
		"generateImport":        generateImport,
		"convertClientRequest":  file.convertClientRequest,
		"convertClientResponse": file.convertClientResponse,
	})

	parsed, err := t.Parse(ServiceClientTemplate)
	if err != nil {
		log.Error(err)
		return
	}
	parsed.Execute(wr, file)
}

func (file *File) convertClientRequest(field Field, expr string) string {
	if field.VariableCall == "" {
		field.VariableCall = fmt.Sprintf("%s.%s", "serviceRequest", field.Variable)
	}

	isRepeated := strings.Contains(field.ProtoType, "repeated")
	if isRepeated {
		field.ProtoType = strings.TrimPrefix(field.ProtoType, "repeated ")
	}

	fieldName := field.Name
	if expr != "" {
		fieldName = expr + "." + field.FieldName
	}

	if ProtoType, ok := protoType2RPCType[field.ProtoType]; ok {
		if isRepeated {
			return fmt.Sprintf(`%s: func() []%s {
					list := make([]%s, len(%s))
					for key, val := range %s {
						list[key] = val
					}
					return list
				}(),`,
				field.Variable,
				ProtoType,
				ProtoType,
				fieldName,
				fieldName,
			)
		} else {
			return fmt.Sprintf("%s: %s(%s),", field.Variable, ProtoType, fieldName)
		}
	}
	switch field.ProtoType {
	case "google.protobuf.Value":
		if isRepeated {
			return fmt.Sprintf(`%s: func() []*_struct.Value {
					list := make([]*_struct.Value, len(%s))
					for key, val := range %s {
						list[key] = proto.EncodeInterface2ProtoValue(val)
					}
					return list
				}(),`,
				field.Variable,
				fieldName,
				fieldName,
			)
		} else {
			return fmt.Sprintf("%s: proto.EncodeInterface2ProtoValue(%s),", field.Variable, fieldName)
		}
	default:
		if isRepeated {
			return fmt.Sprintf(`%s: func() []*%s {
					list := make([]*%s, len(%s))
					for key, val := range %s {
						list[key] = &%s{
							%s
						}
					}
					return list
				}(),`,
				field.Variable,
				field.ProtoType,
				field.ProtoType,
				fieldName,
				fieldName,
				field.ProtoType,
				func() string {
					str := strings.Builder{}
					for _, fileStruct := range file.Structs {
						if field.ProtoType == fileStruct.Name {
							for _, structField := range fileStruct.Fields {
								str.WriteString(fmt.Sprintf("%s\n", file.convertClientRequest(structField, "val")))
							}
						}
					}
					return str.String()
				}(),
			)
		} else {
			if field.GoType == "map[string]interface{}" {
				return fmt.Sprintf(`%s: func() map[string]*_struct.Value {
						param := make(map[string]*_struct.Value, len(%s))
						for key, val := range %s {
							param[key] = proto.EncodeInterface2ProtoValue(val)
						}
						return param
					}(),`,
					field.Variable,
					fieldName,
					fieldName,
				)
			} else {
				return fmt.Sprintf(`%s: &%s{
						%s
					},`,
					field.Variable,
					field.ProtoType,
					func() string {
						str := strings.Builder{}
						for _, fileStruct := range file.Structs {
							if field.ProtoType == fileStruct.Name {
								for _, structField := range fileStruct.Fields {
									str.WriteString(fmt.Sprintf("%s\n", file.convertClientRequest(structField, fieldName)))
								}
							}
						}
						return str.String()
					}(),
				)
			}
		}
	}
	return field.Name
}

func (file *File) convertClientResponse(field Field, expr string) string {
	if field.VariableCall == "" {
		field.VariableCall = field.Name
	}

	repeated := strings.Contains(field.ProtoType, "repeated")
	if repeated {
		field.ProtoType = strings.TrimPrefix(field.ProtoType, "repeated ")
	}

	if expr != "" {
		field.Variable = expr + "." + field.Variable
	}

	if _, ok := protoType2RPCType[field.ProtoType]; ok {
		if repeated {
			return fmt.Sprintf(`func() %s {
					list := make(%s, len(%s))
					for key, val := range %s {
						list[key] = val
					}
					return list
				}()`,
				field.GoType,
				field.GoType,
				field.Variable,
				field.Variable,
			)
		} else {
			return fmt.Sprintf("%s(%s)", field.GoType, field.Variable)
		}
	}
	switch field.ProtoType {
	case "google.protobuf.Value":
		if repeated {
			if strings.Contains(field.GoType, "error") {
				return fmt.Sprintf(`func() []error {
						errorSlice := make([]error, len(%s))
						for key, val := range %s {
							errorSlice[key] = fmt.Errorf("%s",%s)
						}
						return errorSlice
					}()`,
					field.Variable,
					field.Variable,
					"%v",
					field.Variable,
				)
			} else {
				return fmt.Sprintf(`func() []interface{} {
						interfaceSlice := make([]interface{}, len(%s))
						for key, val := range %s {
							interfaceSlice[key] = proto.DecodeProtoValue2Interface(val)
						}
						return interfaceSlice
					}()`,
					field.Variable,
					field.Variable,
				)
			}
		} else {
			if strings.Contains(field.GoType, "error") {
				return fmt.Sprintf(`func() error {
						if %s != nil {
							return fmt.Errorf("%s",%s)
						} else {
							return nil
						}
					}()`,
					field.Variable,
					"%v",
					field.Variable,
				)
			} else {
				return fmt.Sprintf(`func() interface{} {
						if %s != nil {
							return proto.DecodeProtoValue2Interface(%s)
						} else {
							return nil
						}
					}()`,
					field.Variable,
					field.Variable,
				)
			}
		}
	default:
		if repeated {
			field.GoType = file.parseGolangStructType(field)
			return fmt.Sprintf(`func() %s {
					list := make(%s, len(%s))
					for key, val := range %s {
						list[key] = %s{
							%s
						}
					}
					return list
				}()`,
				field.GoType,
				field.GoType,
				field.Variable,
				field.Variable,
				strings.Replace(strings.TrimPrefix(field.GoType, "[]"), "*", "&", 1),
				func() string {
					str := strings.Builder{}
					for _, fileStruct := range file.Structs {
						if field.ProtoType == fileStruct.Name {
							for _, structField := range fileStruct.Fields {
								str.WriteString(fmt.Sprintf("%s: %s,\n", structField.FieldName, file.convertClientResponse(structField, "val")))
							}
						}
					}
					return str.String()
				}(),
			)
		} else {
			if field.GoType == "map[string]interface{}" {
				return fmt.Sprintf(`func() map[string]interface{} {
						param := make(map[string]interface{}, len(%s))
						for key, val := range %s {
							param[key] = proto.DecodeProtoValue2Interface(val)
						}
						return param
					}()`,
					field.Variable,
					field.Variable,
				)
			} else {
				return fmt.Sprintf(`%s{
						%s
					}`,
					strings.Replace(field.GoType, "*", "&", 1),
					func() string {
						str := strings.Builder{}
						for _, fileStruct := range file.Structs {
							if field.ProtoType == fileStruct.Name {
								for _, structField := range fileStruct.Fields {
									str.WriteString(fmt.Sprintf("%s: %s,\n", structField.FieldName, file.convertClientResponse(structField, field.Variable)))
								}
							}
						}
						return str.String()
					}(),
				)
			}
		}
	}
	return field.Name
}
