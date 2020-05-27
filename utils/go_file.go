package utils

import (
	"strings"
)

func (file *File) ParseProtoType(golangType string) string {
	builder := strings.Builder{}
	if strings.HasPrefix(golangType, "[]") {
		if strings.HasPrefix(golangType, "[]byte") {
			return "bytes"
		} else {
			builder.WriteString("repeated ")

			golangType = strings.TrimPrefix(golangType, "[]")
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

					if golangType != "context.Context" {
						// file.Message[protoType] = golangType
						// key: protoType; val: golangType
						if structImportPath, ok := file.ImportsA[protoType[:index]]; ok {
							file.initStructMessage(structImportPath)
							if !file.checkStructMessage(structImportPath, protoType) {
								file.StructMessage[structImportPath] = append(
									file.StructMessage[structImportPath],
									Message{
										Name:     protoType,
										ExprName: golangType,
										FullName: structImportPath,
									},
								)
							}
						}
					}
				} else {
					builder.WriteString(protoType)

				}
			}
		}
	}

	return builder.String()
}

func (file *File) checkImports() {
	for importPath := range file.ImportsA {
		if _, ok := file.StructMessage[file.ImportsA[importPath]]; ok {
			continue
		}
		delete(file.ImportsA, importPath)
	}
}

func (file *File) checkStructMessage(PackagePath, MessageName string) bool {
	for _, StructMessage := range file.StructMessage[PackagePath] {
		if StructMessage.Name == MessageName {
			return true
		}
	}
	return false
}

func (file *File) initStructMessage(StructImportPath string) {
	if file.StructMessage == nil {
		file.StructMessage = make(map[string][]Message, 0)
	}
	if _, ok := file.StructMessage[StructImportPath]; !ok {
		file.StructMessage[StructImportPath] = make([]Message, 0)
	}
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
