package utils

import (
	"fmt"
	"strings"
)

var (
	golangBaseType2ProtoBaseType = map[string]string{
		"bool":    "bool",
		"string":  "string",
		"int":     "sint64",
		"int32":   "sint64",
		"int64":   "sint64",
		"uint":    "uint64",
		"uint32":  "uint64",
		"uint64":  "uint64",
		"float32": "double",
		"float64": "double",
	}
	golangType2ProtoType = map[string]string{
		"error":       "google.protobuf.Value",
		"interface{}": "google.protobuf.Value",
	}
	golangType2JsonType = map[string]string{
		"byte":    "Number",
		"int":     "Number",
		"int32":   "Numner",
		"int64":   "Number",
		"uint":    "Number",
		"uint32":  "Numner",
		"uint64":  "Number",
		"float32": "Number",
		"float64": "Number",
		"string":  "String",
		"bool":    "Boolean",
	}
	golangBaseType = map[string]struct{}{
		"byte":    {},
		"bool":    {},
		"string":  {},
		"int":     {},
		"int32":   {},
		"int64":   {},
		"uint":    {},
		"uint32":  {},
		"uint64":  {},
		"float32": {},
		"float64": {},

		"error":       {},
		"interface{}": {},
	}
)

type Package struct {
	Name         string
	PackagePath  string
	Files        []File
	MessageTypes map[string][]string
	root         *Package
}

type File struct {
	Name          string                        // 文件名
	PackagePath   string                        // 文件包路径
	ProjectPath   string                        // 项目路径 go.mod.module
	GenProtoPath  string                        // 生成 proto 文件路径
	GenInterPath  string                        // 生成 controllers 路径
	GenLogicPath  string                        // 逻辑层包路径 Func 首单词分组
	ProtoPackage  string                        // 生成 proto 文件 package 名
	Structs       map[string]map[string][]Field // [package][structname]fields
	Interfaces    []Interface
	ImportA       map[string]string
	ImportB       map[string]string
	Message       map[string]string
	StructMessage map[string][]Message
}

type Interface struct {
	Funcs []Func
	Name  string
}

type Func struct {
	Name    string
	Group   string
	Params  []Field
	Results []Field
}

type Field struct {
	Name      string // 字段名 原参数名或返回值名或struct中的字段名
	GoType    string // go类型
	JsonType  string // json 类型
	ProtoType string // proto类型
	Comment   string // 注释
	GoExpr    string // go类型的引用前缀
	Package   string // go类型定义的所在包
}

type Message struct {
	Name     string // struct名字
	ExprName string // 调用名 （pager.PagerListResp）
	FullName string // 全名 （带包名）
}

func title(name string) string {
	builder := strings.Builder{}
	for _, val := range strings.Split(name, "_") {
		builder.WriteString(strings.Title(val))
	}
	return builder.String()
}

func generateImport(key, val string) string {
	sort := packageSort(val)
	if key == sort {
		return fmt.Sprintf(`"%s"`, val)
	} else {
		return fmt.Sprintf(`%s "%s"`, key, val)
	}
}

func packageSort(Package string) string {
	if index := strings.LastIndex(Package, "/"); index != -1 {
		return Package[index+1:]
	} else {
		return Package
	}
}

func parseJsonType(fieldType string) string {
	jsonType := strings.Builder{}
	fieldType = strings.TrimPrefix(fieldType, "*")
	if strings.HasPrefix(fieldType, "[]") {
		fieldType = strings.TrimPrefix(fieldType, "[]")
		jsonType.WriteString("[]")
	}
	if val, ok := golangType2JsonType[fieldType]; ok {
		jsonType.WriteString(val)
	} else {
		jsonType.WriteString("Object")
	}
	return jsonType.String()
}

func merge(a, b map[string][]Field) map[string][]Field {
	for key, val := range b {
		a[key] = val
	}
	return a
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
