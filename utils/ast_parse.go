package utils

import (
	"fmt"
	"go/ast"
	"go/parser"
	"strconv"
	"strings"

	log "github.com/cihub/seelog"
	"golang.org/x/tools/go/loader"
)

func (file *File) ParseFile(astFile *ast.File) {
	file.ParseImport(astFile)
	inters := make([]Interface, 0)
	ast.Inspect(astFile, func(x ast.Node) bool {
		inter := Interface{}
		switch x.(type) {
		case *ast.TypeSpec:
			typeSpec := x.(*ast.TypeSpec)
			inter.Name = typeSpec.Name.Name
			interfaceType, ok := typeSpec.Type.(*ast.InterfaceType)
			if !ok {
				return true
			}
			log.Info("find interface: ", inter.Name)
			inter.Funcs = make([]Func, len(interfaceType.Methods.List))
			for index, field := range interfaceType.Methods.List {
				fun := file.ParseFunc(field.Names[0].Name, field.Type.(*ast.FuncType))
				inter.Funcs[index] = fun
			}
			inters = append(inters, inter)
		case *ast.FuncDecl:
			decl := x.(*ast.FuncDecl)
			if decl.Recv != nil {
				return true
			}
			if decl.Name.Name[0] != strings.ToUpper(decl.Name.Name)[0] {
				return true
			}
			inter.Name = decl.Name.Name + "Func"
			log.Info("find func: ", inter.Name)
			funcType := decl.Type
			fun := file.ParseFunc(decl.Name.Name, funcType)
			inter.Funcs = []Func{fun}
			inters = append(inters, inter)
		default:
			return true
		}
		return false
	})
	file.Interfaces = inters
}

func (file *File) ParseImport(astFile *ast.File) {
	ImportA := make(map[string]string)
	ast.Inspect(astFile, func(x ast.Node) bool {
		switch x.(type) {
		case *ast.ImportSpec:
			importSpec := x.(*ast.ImportSpec)
			var key string
			val := importSpec.Path.Value
			val, _ = strconv.Unquote(val)
			if importSpec.Name != nil {
				key = importSpec.Name.Name
			} else {
				lastIndex := strings.LastIndex(val, "/")
				if lastIndex == -1 {
					key = val
				} else {
					key = val[lastIndex+1:]
				}
			}
			ImportA[key] = val
		default:
			return true
		}
		return false
	})
	ImportA["context"] = "context"
	file.ImportA = ImportA
	file.ImportB = make(map[string]string, 0)
	for key, val := range file.ImportA {
		file.ImportB[val] = key
	}
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

	if funcType.Params != nil {
		fun.Params = file.ParseField(funcType.Params.List)
	}
	if funcType.Results != nil {
		fun.Results = file.ParseField(funcType.Results.List)
	}

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
	fields := make([]Field, 0)
	for _, field := range astField {
		fieldType := parseExpr(field.Type)
		protoType := file.parseType(fieldType)
		if field.Names == nil {
			conf := loader.Config{ParserMode: parser.ParseComments}
			conf.Import(file.PackagePath)
			program, err := conf.Load()
			if err != nil {
				log.Error(err)
				continue
			}
			astFiles := program.Package(file.PackagePath).Files
			Root := Package{
				PackagePath: file.PackagePath,
				Files:       make([]File, 0, len(astFiles)),
				root:        &Package{MessageTypes: map[string][]string{}},
			}
			for _, astFile := range astFiles {
				structFile := Root.ParseStruct([]Message{{
					Name:     fieldType,
					ExprName: fieldType,
					FullName: file.PackagePath,
				}}, astFile)
				for _, x := range structFile.Structs {
					if field, ok := x[fieldType]; ok {
						fields = append(fields, field...)
					}
				}
			}
		}

		expr, packageImport := func() (expr string, packageImport string) {
			name := strings.TrimPrefix(strings.TrimPrefix(fieldType, "[]"), "*")
			if index := strings.Index(name, "."); index != -1 {
				expr = name[:index]
			}

			if _, ok := golangBaseType[name]; !ok {
				if importA, ok := file.ImportA[expr]; ok {
					packageImport = importA
				} else {
					packageImport = file.PackagePath
				}
			}
			return
		}()

		for _, value := range field.Names {
			fields = append(fields,
				Field{
					Name:      value.Name,
					GoType:    fieldType,
					JsonType:  parseJsonType(fieldType),
					ProtoType: protoType,
					Comment:   parseComment(field),
					GoExpr:    expr,
					Package:   packageImport,
				},
			)
		}
	}
	return fields
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

func parseComment(field *ast.Field) string {
	if field.Comment != nil {
		for _, comment := range field.Comment.List {
			return strings.TrimSpace(strings.TrimPrefix(comment.Text, "//"))
		}
	}
	return ""
}
