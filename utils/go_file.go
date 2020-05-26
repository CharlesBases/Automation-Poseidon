package utils

import (
	"go/parser"
	"path"
	"strings"

	"golang.org/x/tools/go/loader"
)

func (file *File) parsePackageStruct(root *Package) {
	file.ParseStructMessage()

	for key, value := range file.StructMessage {
		Info("parse structs in package: ", key)
		conf := loader.Config{ParserMode: parser.ParseComments}
		conf.Import(key)
		program, err := conf.Load()
		ThrowCheck(err)
		astFiles := program.Package(key).Files
		Root := &Package{
			Name:        path.Base(key),
			PackagePath: key,
			Files:       make([]File, 0),
			root:        root,
		}
		for _, astFile := range astFiles {
			structFile := Root.ParseStruct(value, astFile)
			if structFile == nil {
				continue
			}
			structFile.parsePackageStruct(root)
			Root.Files = append(Root.Files, *structFile)
		}
		if len(Root.Files) == 0 {
			continue
		}
	}

	// 合并结果
	// for packageValue := range packageschannel {
	// 	for _, fileValue := range packageValue.Files {
	// 		file.Structs[packageValue.PackagePath] = merge(fileValue.Structs[packageValue.PackagePath], file.Structs[packageValue.PackagePath])
	// 		for key, val := range fileValue.ImportsA {
	// 			file.ImportsA[key] = val
	// 		}
	// 	}
	// }
	//
	// file.ImportsB = map_conversion(file.ImportsA)
}

func (file *File) ParseStructMessage() {
	structMessage := make(map[string][]Message, 0)
	for key, val := range file.Message {
		imp := strings.TrimPrefix(val, "*")
		if index := strings.Index(imp, "."); index != -1 {
			impPrefix := imp[:index]
			imp, ok := file.ImportsA[impPrefix]
			if ok {
				if structMessage[imp] == nil {
					structMessage[imp] = make([]Message, 0)
				}
				structMessage[imp] = append(
					structMessage[imp],
					Message{
						Name:     key,
						ExprName: val,
						FullName: imp,
					},
				)
			}
		} else {
			_, ok := golangBaseType[val]
			if !ok {
				pkgpath := file.PackagePath
				if structMessage[pkgpath] == nil {
					structMessage[pkgpath] = make([]Message, 0)
				}
				message := Message{
					Name:     key,
					ExprName: val,
					FullName: pkgpath,
				}
				structMessage[pkgpath] = append(structMessage[pkgpath], message)
			}
		}
	}
	file.StructMessage = structMessage
}

func (file *File) parseProtoType(golangType string) string {
	if file.Message == nil {
		file.Message = make(map[string]string, 0)
	}
	builder := strings.Builder{}
	if strings.HasPrefix(golangType, "[]") {
		if strings.HasPrefix(golangType, "[]byte") {
			return "bytes"
		} else {
			golangType = strings.TrimPrefix(golangType, "[]")
			builder.WriteString("repeated ")
		}
	}
	if protoBaseType, ok := golangBaseType2ProtoBaseType[golangType]; ok {
		builder.WriteString(protoBaseType)
	} else {
		if protoType, ok := golangType2ProtoType[golangType]; ok {
			builder.WriteString(protoType)
		} else {
			if strings.HasPrefix(golangType, "map") {
				if index := strings.Index(golangType, "]"); index != -1 {
					builder.WriteString("map<string, google.protobuf.Value>")
				}
			} else {
				protoType = strings.TrimPrefix(golangType, "*")
				if index := strings.LastIndex(protoType, "."); index != -1 {
					builder.WriteString(protoType[index+1:])

					typePrefix := protoType[:index]
					if imp, ok := file.ImportsA[typePrefix]; ok {
						if file.StructMessage[imp] == nil {
							file.StructMessage[imp] = make([]Message, 0)
						}
						file.StructMessage[imp] = append(
							file.StructMessage[imp],
							Message{
								Name:     protoType[index+1:],
								ExprName: golangType,
								FullName: imp,
							})
					}
				} else {
					builder.WriteString(protoType)

				}
			}
		}
	}

	return builder.String()
}

func (file *File) GoTypeConfig() {
	for interfaceKey, interfaceValue := range file.Interfaces {
		for funcKey, funcValue := range interfaceValue.Funcs {
			for resultKey, resultVallue := range funcValue.Results {
				goType := file.typeConfig(&resultVallue)
				file.Interfaces[interfaceKey].Funcs[funcKey].Results[resultKey].GoType = goType
			}
			for paramKey, paramValue := range funcValue.Params {
				goType := file.typeConfig(&paramValue)
				file.Interfaces[interfaceKey].Funcs[funcKey].Params[paramKey].GoType = goType
			}
		}
	}
	for packageName, packageStruct := range file.Structs {
		for structName, fields := range packageStruct {
			for fieldIndex, field := range fields {
				goType := file.typeConfig(&field)
				file.Structs[packageName][structName][fieldIndex].GoType = goType
			}
		}
	}
}

func (file *File) typeConfig(field *Field) string {
	if ImportA, ok := file.ImportsA[field.Package]; ok {
		i := 0
		if strings.Contains(field.GoType, "[]") {
			i = i + 2
		}
		if strings.Contains(field.GoType, "*") {
			i = i + 1
		}
		index := strings.Index(field.GoType, ".")
		if i == 0 {
			if index == -1 {
				return ImportA + "." + field.GoType
			}
			return ImportA + "." + field.GoType[index+1:]
		}
		if index == -1 {
			return field.GoType[:i] + ImportA + "." + field.GoType[i:]
		}
		return field.GoType[:i] + ImportA + "." + field.GoType[index+1:]
	}
	return field.GoType
}
