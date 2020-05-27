package utils

import (
	"go/ast"
	"go/parser"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"golang.org/x/tools/go/loader"
)

func (file *File) ParseFile(astFile *ast.File) {
	swg := sync.WaitGroup{}
	swg.Add(2)

	go func() {
		defer swg.Done()

		file.ParseImport(astFile)
	}()

	go func() {
		defer swg.Done()

		file.ParseInterface(astFile)
	}()

	swg.Wait()

	file.ParsePackageStruct(
		&Package{
			PackagePath: file.PackagePath,
		},
	)
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

	file.ImportsA = import_merge(file.ImportsA, imports)
}

func (file *File) ParseInterface(astFile *ast.File) {
	inters := make([]Interface, 0)

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
					swg := sync.WaitGroup{}
					inter.Funcs = make([]Func, len(interfaceType.Methods.List))
					for index := range interfaceType.Methods.List {
						swg.Add(1)
						go func(num int, field *ast.Field) {
							defer swg.Done()
							inter.Funcs[num] = file.ParseFunc(field.Names[0].Name, field.Type.(*ast.FuncType))
						}(index, interfaceType.Methods.List[index])
					}
					swg.Wait()
					inters = append(inters, inter)
				}
			}
			return true
		},
	)

	file.Interfaces = inters
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

		swg.Add(2)

		go func() {
			defer swg.Done()

			root.root = Root
			root.Name = path.Base(key)
			root.PackagePath = key
			root.Files = make([]*File, 0)
		}()

		go func() {
			defer swg.Done()

			conf := loader.Config{ParserMode: parser.ParseComments}
			conf.Import(key)
			program, err := conf.Load()
			ThrowCheck(err)
			astFiles = program.Package(key).Files
		}()

		swg.Wait()

		for index := range astFiles {
			swg.Add(1)

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
			file.Structs = struct_merge(file.Structs, fileValue.Structs)
			file.ImportsA = import_merge(file.ImportsA, fileValue.ImportsA)
		}
	}
}

func (root *Package) ParseStruct(messages []Message, astFile *ast.File) *File {
	file := File{
		PackagePath: root.PackagePath,
		Structs:     make(map[string]map[string][]Field, 0),
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
						if root.root.MessageTypes != nil {
							if messageTypes, ok := root.root.MessageTypes[root.PackagePath]; ok {
								for messageTypesIndex := range messageTypes {
									if strings.HasSuffix(messageTypes[messageTypesIndex], typeSpec.Name.Name) {
										return true
									}
								}
							} else {
								root.root.MessageTypes[root.PackagePath] = make([]string, 0)
							}
						} else {
							root.root.MessageTypes = make(map[string][]string, 0)
						}
						Info("find struct: ", typeSpec.Name.Name)
						structs[typeSpec.Name.Name] = file.ParseStruct(typeSpec.Name.Name, structType)
						root.root.MessageTypes[root.PackagePath] = append(root.root.MessageTypes[root.PackagePath], typeSpec.Name.Name)

						break
					}
				}
			}
		}
		return true
	})

	file.Structs[file.PackagePath] = structs
	file.checkImports()
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
	/*
		swg := sync.WaitGroup{}
		fields := make([]Field, len(astField))

		for index := range astField {
			swg.Add(1)

			go func(number int, field *ast.Field) {
				defer swg.Done()

				fieldType := parseExpr(field.Type)

				expr, packageImport := func() (expr string, packageImport string) {
					name := strings.TrimPrefix(strings.TrimPrefix(fieldType, "[]"), "*")
					if index := strings.Index(name, "."); index != -1 {
						expr = name[:index]
					}

					if _, ok := golangBaseType[name]; !ok {
						if imp, ok := file.ImportsA[expr]; ok {
							packageImport = imp
						} else {
							packageImport = file.PackagePath
						}
					}
					return
				}()

				for _, value := range field.Names {
					fields[number] = Field{
						Name:      value.Name,
						GoType:    fieldType,
						JsonType:  parseJsonType(fieldType),
						ProtoType: file.parseProtoType(fieldType),
						Comment:   parseComment(field),
						GoExpr:    expr,
						Package:   packageImport,
					}
				}
			}(index, astField[index])
		}

		swg.Wait()
	*/

	fields := make([]Field, len(astField))
	for index := range astField {
		fieldType := ParseExpr(astField[index].Type)

		expr, packageImport := func() (expr string, packageImport string) {
			name := strings.TrimPrefix(strings.TrimPrefix(fieldType, "[]"), "*")
			if index := strings.Index(name, "."); index != -1 {
				expr = name[:index]
			}

			if _, ok := golangBaseType[name]; !ok {
				if imp, ok := file.ImportsA[expr]; ok {
					packageImport = imp
				} else {
					packageImport = file.PackagePath
				}
			}
			return
		}()

		for _, value := range astField[index].Names {
			fields[index] = Field{
				Name:      value.Name,
				GoType:    fieldType,
				JsonType:  parseJsonType(fieldType),
				ProtoType: file.ParseProtoType(fieldType),
				Comment:   parseComment(astField[index]),
				GoExpr:    expr,
				Package:   packageImport,
			}
		}
	}

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
