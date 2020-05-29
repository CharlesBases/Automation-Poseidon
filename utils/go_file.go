package utils

import (
	"strings"
)

func (file *File) ParseProtoType(golangType string) string {
	builder := strings.Builder{}
	golangType = strings.TrimPrefix(golangType, "*")
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

					if structImportPath, ok := file.Imports[protoType[:index]]; ok {
						file.mutex.Lock()
						if _, ok := file.StructMessage[structImportPath]; !ok {
							file.StructMessage[structImportPath] = make([]Message, 0)
						}
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
						file.mutex.Unlock()
					}
				} else {
					builder.WriteString(protoType)
				}
			}
		}
	}

	return builder.String()
}

func (file *File) merge_structs(target map[string]Struct) {
	if file.Structs == nil || len(file.Structs) == 0 {
		file.Structs = target
		return
	}
	for packagePath, StructInfo := range target {
		if _, ok := file.Structs[packagePath]; ok {
			for structName, structFields := range StructInfo {
				if _, ok := file.Structs[structName]; !ok {
					file.Structs[packagePath][structName] = structFields
				}
			}
		} else {
			file.Structs[packagePath] = StructInfo
		}
	}
	return
}

func (file *File) merge_imports(target map[string]string) {
	if file.Imports == nil || len(file.Imports) == 0 {
		return
	}
	for key, val := range target {
		if _, ok := file.Imports[key]; !ok {
			file.Imports[key] = val
		}
	}
	return
}

func (file *File) checkImports() {
	for importPath := range file.Imports {
		if _, ok := file.StructMessage[file.Imports[importPath]]; ok {
			continue
		}
		delete(file.Imports, importPath)
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

func parseJsonType(fieldType string) string {
	jsonType := strings.Builder{}
	fieldType = strings.TrimPrefix(fieldType, "*")
	if strings.HasPrefix(fieldType, "[]") {
		jsonType.WriteString("[]")
		fieldType = strings.TrimPrefix(fieldType, "[]")
	}
	fieldType = strings.TrimPrefix(fieldType, "*")
	if val, ok := golangType2JsonType[fieldType]; ok {
		jsonType.WriteString(val)
	} else {
		jsonType.WriteString("Object")
	}
	return jsonType.String()
}
