package utils

import (
	"go/ast"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
)

func (root *Package) ParseStruct(message []Message, astFile *ast.File) *File {
	file := File{
		PackagePath: root.PackagePath,
		Structs:     make(map[string]map[string][]Field, 0),
	}

	file.parseImport(astFile)

	structs := make(map[string][]Field, 0)
	ast.Inspect(astFile, func(x ast.Node) bool {
		switch x.(type) {
		case *ast.TypeSpec:
			typeSpec := x.(*ast.TypeSpec)

			if structType, ok := typeSpec.Type.(*ast.StructType); ok {
				var (
					isContainsA bool
					isContainsB bool
				)
				if message == nil {
					isContainsA = true
				} else {
					for _, v := range message {
						if v.Name == typeSpec.Name.Name {
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
							if v == typeSpec.Name.Name {
								isContainsB = true
							}
						}
					} else {
						root.root.MessageTypes[root.PackagePath] = make([]string, 0)
					}
				}
				if isContainsA && !isContainsB {
					fields := file.ParseStruct(typeSpec.Name.Name, structType)
					Info("find struct: ", typeSpec.Name.Name)
					structs[typeSpec.Name.Name] = fields
					root.root.MessageTypes[root.PackagePath] = append(root.root.MessageTypes[root.PackagePath], typeSpec.Name.Name)
				}
			}
		}
		return true
	})
	file.Structs[file.PackagePath] = structs
	return &file
}

func (file *File) ParseFile(astFile *ast.File) {
	swg := sync.WaitGroup{}
	swg.Add(2)

	go func() {
		defer swg.Done()

		file.parseImport(astFile)
	}()

	go func() {
		defer swg.Done()

		file.parseInterface(astFile)
	}()

	swg.Wait()

	file.parsePackageStruct(
		&Package{
			PackagePath: file.PackagePath,
		},
	)
}

func (file *File) parseImport(astFile *ast.File) {
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

	map_merge(file.ImportsA, imports)
}

func (file *File) parseInterface(astFile *ast.File) {
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

// 解析ast方法声明中的表达式
func parseExpr(expr ast.Expr) (fieldType string) {
	switch expr.(type) {
	case *ast.StarExpr:
		starExpr := expr.(*ast.StarExpr)
		return "*" + parseExpr(starExpr.X)
	case *ast.SelectorExpr:
		selectorExpr := expr.(*ast.SelectorExpr)
		return parseExpr(selectorExpr.X) + "." + selectorExpr.Sel.Name
	case *ast.ArrayType:
		arrayType := expr.(*ast.ArrayType)
		return "[]" + parseExpr(arrayType.Elt)
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

func (file *File) ParseField(astField []*ast.Field) []Field {
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

	return fields
}

func parseComment(field *ast.Field) string {
	if field.Comment != nil {
		for _, comment := range field.Comment.List {
			return strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		}
	}
	return ""
}
