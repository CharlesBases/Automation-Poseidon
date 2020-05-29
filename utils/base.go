package utils

import (
	"os"
	"path/filepath"
	"strings"
	"sync"
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

var (
	Poseidon = new(Information)

	basefile = new(File)
	baseroot = new(Package)

	importsMutex = sync.RWMutex{}
	structsMutex = sync.RWMutex{}
)

type Struct map[string][]Field // [structname]fields

type Information struct {
	PackagePath  string
	ProjectPath  string
	GenProtoPath string
	GenInterPath string
	GenLogicPath string
	ProtoPackage string

	Imports    map[string]string
	Structs    map[string]Struct
	Interfaces []Interface
}

type (
	Package struct {
		PackagePath    string
		Files          []*File
		StructMessages map[string][]string

		mutex sync.RWMutex
		root  *Package
	}

	File struct {
		PackagePath   string            // 文件包路径
		Structs       map[string]Struct // [package][structname]fields
		Interfaces    []Interface
		Imports       map[string]string
		ImportsA      map[string]string
		ImportsB      map[string]string
		StructMessage map[string][]Message

		mutex sync.RWMutex
	}

	Interface struct {
		Funcs []Func
		Name  string
	}

	Func struct {
		Name    string
		Group   string
		Params  []Field
		Results []Field
	}

	Field struct {
		Name         string // 字段名
		VariableType string // 字段类型
		JsonType     string // json 类型
		ProtoType    string // proto 类型
		Comment      string // 注释

		Package string // go类型定义的所在包
		Prefix  string // go类型的引用前缀
		Type    string
	}

	Message struct {
		Name     string // struct名字
		ExprName string // 调用名
		FullName string // 结构体所在包路径
	}
)

func (*Information) Init(src, SourceFilePath, ProjectPath, GenInterPath, GenLogicPath, GenProtoPath, ProtoPackage string) {
	Poseidon.Imports = make(map[string]string, 0)
	Poseidon.Structs = make(map[string]Struct, 0)

	Poseidon.ProtoPackage = ProtoPackage
	Poseidon.PackagePath = strings.TrimPrefix(filepath.Dir(SourceFilePath), src)[1:]

	Poseidon.GenInterPath = Poseidon.parsepath(src, GenInterPath)
	Poseidon.GenLogicPath = Poseidon.parsepath(src, GenLogicPath)
	Poseidon.GenProtoPath = Poseidon.parsepath(src, GenProtoPath)

	Poseidon.ProjectPath = func() string {
		if len(ProjectPath) != 0 {
			return ProjectPath
		}
		return filepath.Dir(Poseidon.PackagePath)
	}()

	basefile.StructMessage = make(map[string][]Message, 0)
	basefile.PackagePath = Poseidon.PackagePath
	baseroot.PackagePath = Poseidon.PackagePath
}

func (*Information) parsepath(src, path string) string {
	abspath, err := filepath.Abs(path)
	ThrowCheck(err)
	os.MkdirAll(abspath, 0755)
	return strings.TrimPrefix(abspath, src)[1:]
}

func (*Information) merge_imports(imports map[string]string) {
	defer importsMutex.Unlock()
	importsMutex.Lock()

	if len(Poseidon.Imports) == 0 {
		Poseidon.Imports = imports
		return
	}

	for key, val := range imports {
		Poseidon.Imports[key] = val
	}
}

func (*Information) merge_structs(structs map[string]Struct) {
	defer structsMutex.RUnlock()
	structsMutex.RLock()

	if len(Poseidon.Structs) == 0 {
		Poseidon.Structs = structs
		return
	}

	for packagePath, structInfor := range structs {
		if _, ok := Poseidon.Structs[packagePath]; ok {
			for structName, structFields := range structInfor {
				if _, ok := Poseidon.Structs[packagePath][structName]; !ok {
					Poseidon.Structs[packagePath][structName] = structFields
				}
			}
		} else {
			Poseidon.Structs[packagePath] = structInfor
		}
	}
}

func (*Information) check_imports() {
	for key, val := range Poseidon.Imports {
		if val != "context" {
			if _, ok := Poseidon.Structs[val]; !ok {
				delete(Poseidon.Imports, key)
			}
		}
	}
}

func (*Information) sortout() {
	// struct
	for Package, Structs := range Poseidon.Structs {
		for structname := range Structs {
			for index := 0; index < len(Poseidon.Structs[Package][structname]); index++ {
				field := Poseidon.Structs[Package][structname][index]
				if len(field.Name) == 0 {
					fields := Poseidon.Structs[Package][structname]

					Poseidon.Structs[Package][structname] = append(fields[:index], Poseidon.Structs[field.Package][field.Type]...)
					Poseidon.Structs[Package][structname] = append(Poseidon.Structs[Package][structname], fields[index+1:]...)
				}
			}
		}
	}
}
