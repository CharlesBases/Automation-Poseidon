package utils

import (
	"go/ast"
	"strings"

	log "github.com/cihub/seelog"
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
	jsonDefaultValue = map[string]interface{}{
		"Number":  0,
		"String":  "",
		"Boolean": false,
		"Object":  nil,
	}
)

type Package struct {
	Name         string
	PackagePath  string
	Files        []File
	MessageTypes map[string][]string
	root         *Package
}

type Message struct {
	Name     string // struct名字
	ExprName string // 调用名 （pager.PagerListResp）
	FullName string // 全名 （带包名）
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

func (root *Package) ParseStruct(message []Message, astFile *ast.File) *File {
	file := File{
		PackagePath: root.PackagePath,
		Structs:     make(map[string]map[string][]Field, 0),
	}

	file.ParseImport(astFile)

	structs := make(map[string][]Field, 0)
	ast.Inspect(astFile, func(x ast.Node) bool {
		switch x.(type) {
		case *ast.TypeSpec:
			spec := x.(*ast.TypeSpec)
			structType, ok := spec.Type.(*ast.StructType)
			if !ok {
				return true
			}
			var (
				isContainsA bool
				isContainsB bool
			)
			if message == nil {
				isContainsA = true
			} else {
				for _, v := range message {
					if v.Name == spec.Name.Name {
						isContainsA = true
					}
				}
			}
			if root.root.MessageTypes == nil {
				root.root.MessageTypes = make(map[string][]string, 0)
				isContainsB = false
			} else {
				messageType, ok := root.root.MessageTypes[root.PackagePath]
				if ok {
					for _, v := range messageType {
						if v == spec.Name.Name {
							isContainsB = true
						}
					}
				} else {
					root.root.MessageTypes[root.PackagePath] = make([]string, 0)
				}
			}
			if isContainsA && !isContainsB {
				fields := file.ParseStruct(spec.Name.Name, structType)
				log.Info("find struct: ", spec.Name.Name)
				structs[spec.Name.Name] = fields
				root.root.MessageTypes[root.PackagePath] = append(root.root.MessageTypes[root.PackagePath], spec.Name.Name)
			}
		default:
			return true
		}
		return false
	})
	file.Structs[file.PackagePath] = structs
	return &file
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
