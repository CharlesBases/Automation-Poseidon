package utils

import (
	"go/ast"
	"go/parser"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/tools/go/loader"
)

func ParseFile(astFile *ast.File) {
	defer Poseidon.sortout()

	swg := sync.WaitGroup{}
	swg.Add(2)

	go func() {
		defer swg.Done()

		basefile.ParseImport(astFile)
	}()

	go func() {
		defer swg.Done()

		basefile.ParseInterface(astFile)
	}()

	swg.Wait()

	basefile.ParsePackageStruct(baseroot)
}

func (file *File) ParseImport(astFile *ast.File) {
	imports := make(map[string]string, 0)

	ast.Inspect(
		astFile,
		func(x ast.Node) bool {
			switch x.(type) {
			case *ast.ImportSpec:
				importSpec := x.(*ast.ImportSpec)
				var (
					key, val string
				)
				val, _ = strconv.Unquote(importSpec.Path.Value)
				if importSpec.Name != nil {
					key = importSpec.Name.Name
				} else {
					key = filepath.Base(val)
				}
				imports[key] = val
			}
			return true
		},
	)

	file.Imports = imports
	Poseidon.merge_imports(imports)
}

func (file *File) ParseInterface(astFile *ast.File) {
	interfaces := make([]Interface, 0)

	ast.Inspect(
		astFile,
		func(x ast.Node) bool {
			inter := Interface{}
			switch x.(type) {
			case *ast.TypeSpec:
				typeSpec := x.(*ast.TypeSpec)
				if interfaceType, ok := typeSpec.Type.(*ast.InterfaceType); ok {
					inter.Name = typeSpec.Name.Name
					Info("find interface: ", inter.Name)
					var (
						swg    = sync.WaitGroup{}
						amount = len(interfaceType.Methods.List)
					)
					swg.Add(amount)
					inter.Funcs = make([]Func, amount)
					for index := range interfaceType.Methods.List {
						go func(num int, field *ast.Field) {
							defer swg.Done()
							inter.Funcs[num] = file.ParseFunc(field.Names[0].Name, field.Type.(*ast.FuncType))
						}(index, interfaceType.Methods.List[index])
					}
					swg.Wait()
					interfaces = append(interfaces, inter)
				}
			}
			return true
		},
	)

	file.Interfaces = interfaces
	Poseidon.Interfaces = interfaces
}

func (file *File) ParsePackageStruct(Root *Package) {
	for key, value := range file.StructMessage {
		Info("parse structs in package: ", key)
		var (
			swg     = sync.WaitGroup{}
			rwMutex = sync.RWMutex{}

			astFiles = make([]*ast.File, 0)
			root     = new(Package)
		)

		root.root = Root
		root.PackagePath = key
		root.Files = make([]*File, 0)

		conf := loader.Config{ParserMode: parser.ParseComments}
		conf.Import(key)
		program, err := conf.Load()
		ThrowCheck(err)
		astFiles = program.Package(key).Files

		swg.Add(len(astFiles))

		for index := range astFiles {
			go func(num int, astfile *ast.File) {
				defer swg.Done()

				structFile := root.ParseStruct(value, astfile)
				if structFile == nil {
					return
				}
				structFile.ParsePackageStruct(Root)

				rwMutex.RLock()
				root.Files = append(root.Files, structFile)
				rwMutex.RUnlock()
			}(index, astFiles[index])
		}

		swg.Wait()

		for _, fileValue := range root.Files {
			file.merge_structs(fileValue.Structs)
			file.merge_imports(fileValue.Imports)
			file.checkImports()

			Poseidon.merge_structs(fileValue.Structs)
			Poseidon.merge_imports(fileValue.Imports)
		}
	}
}

func (root *Package) ParseStruct(messages []Message, astFile *ast.File) *File {
	file := File{
		PackagePath:   root.PackagePath,
		Structs:       make(map[string]Struct, 0),
		StructMessage: make(map[string][]Message, 0),
	}

	file.ParseImport(astFile)

	structs := make(map[string][]Field, 0)
	ast.Inspect(astFile, func(x ast.Node) bool {
		switch x.(type) {
		case *ast.TypeSpec:
			typeSpec := x.(*ast.TypeSpec)

			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				for messagesIndex := range messages {
					// 文件中结构体是否需要
					if strings.HasSuffix(messages[messagesIndex].ExprName, typeSpec.Name.Name) {
						// 文件中结构体是否已加载
						if root.root.StructMessages != nil {
							if messageTypes, ok := root.root.StructMessages[root.PackagePath]; ok {
								for messageTypesIndex := range messageTypes {
									if strings.HasSuffix(messageTypes[messageTypesIndex], typeSpec.Name.Name) {
										return true
									}
								}
							} else {
								root.root.StructMessages[root.PackagePath] = make([]string, 0)
							}
						} else {
							root.root.StructMessages = make(map[string][]string, 0)
						}
						Info("find struct: ", typeSpec.Name.Name)
						structs[typeSpec.Name.Name] = file.ParseStruct(typeSpec.Name.Name, structType)
						root.root.mutex.RLock()
						root.root.StructMessages[root.PackagePath] = append(root.root.StructMessages[root.PackagePath], typeSpec.Name.Name)
						root.root.mutex.RUnlock()
						break
					}
				}
			}
		}
		return true
	})

	file.Structs[file.PackagePath] = structs
	file.checkImports()

	Poseidon.merge_structs(map[string]Struct{file.PackagePath: structs})
	Poseidon.check_imports()
	return &file
}

func (file *File) ParseStruct(name string, structType *ast.StructType) []Field {
	fields := make([]Field, 0)
	for _, field := range file.ParseField(structType.Fields.List) {
		if strings.Title(field.Name) == field.Name {
			fields = append(fields, field)
		}
	}
	return fields
}

// 解析ast函数
func (file *File) ParseFunc(name string, funcType *ast.FuncType) Func {
	fun := Func{
		Name: name,
		Group: func() string {
			group := strings.Builder{}
			for i, x := range []rune(name) {
				if i != 0 && x < 91 && x > 64 {
					break
				}
				group.WriteString(string(x))
			}
			return group.String()
		}(),
	}

	swg := sync.WaitGroup{}
	swg.Add(2)

	go func() {
		defer swg.Done()

		if funcType.Params != nil {
			fun.Params = file.ParseField(funcType.Params.List)
		}
	}()

	go func() {
		defer swg.Done()

		if funcType.Results != nil {
			fun.Results = file.ParseField(funcType.Results.List)
		}
	}()

	swg.Wait()

	return fun
}

func (file *File) ParseField(astField []*ast.Field) []Field {
	var (
		swg    = sync.WaitGroup{}
		fields = make([]Field, len(astField))
	)

	swg.Add(len(astField))

	for index := range astField {
		go func(number int, field *ast.Field) {
			defer swg.Done()

			fieldType := ParseExpr(field.Type)

			var (
				prefix        string
				gotype        string
				packageimport string
			)

			gotype = strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(fieldType, "*"), "[]"), "*")
			if index := strings.Index(gotype, "."); index != -1 {
				prefix = gotype[:index]
				gotype = gotype[index+1:]
			}

			if _, ok := golangBaseType[gotype]; !ok {
				if imp, ok := file.Imports[prefix]; ok {
					packageimport = imp
				} else {
					packageimport = file.PackagePath
				}
			}

			fields[number] = Field{
				VariableType: fieldType,
				JsonType:     parseJsonType(fieldType),
				ProtoType:    file.ParseProtoType(fieldType),
				Comment:      parseComment(field),

				Package: packageimport,
				Prefix:  prefix,
				Type:    gotype,
			}

			if astField[number].Names != nil {
				fields[number].Name = field.Names[0].Name
			}
		}(index, astField[index])
	}

	swg.Wait()

	return fields
}

func ParseExpr(expr ast.Expr) (fieldType string) {
	switch expr.(type) {
	case *ast.StarExpr:
		starExpr := expr.(*ast.StarExpr)
		return "*" + ParseExpr(starExpr.X)
	case *ast.SelectorExpr:
		selectorExpr := expr.(*ast.SelectorExpr)
		return ParseExpr(selectorExpr.X) + "." + selectorExpr.Sel.Name
	case *ast.ArrayType:
		arrayType := expr.(*ast.ArrayType)
		return "[]" + ParseExpr(arrayType.Elt)
	case *ast.MapType:
		return "map[string]interface{}"
	case *ast.InterfaceType:
		return "interface{}"
	case *ast.Ident:
		ident := expr.(*ast.Ident)
		return ident.Name
	default:
		return fieldType
	}
}

func parseComment(field *ast.Field) string {
	if field.Comment != nil {
		for _, comment := range field.Comment.List {
			return strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		}
	}
	return ""
}
