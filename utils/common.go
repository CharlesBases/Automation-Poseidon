package utils

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
	ImportsA      map[string]string
	ImportsB      map[string]string
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
