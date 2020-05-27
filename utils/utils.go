package utils

import (
	"fmt"
	"strings"
)

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

// map[key]value tp map[value]key
func map_conversion(source map[string]string) map[string]string {
	finish := make(map[string]string, 0)
	for key, val := range source {
		finish[val] = key
	}
	return finish
}

func struct_merge(source, target map[string]map[string][]Field) map[string]map[string][]Field {
	if source == nil || len(source) == 0 {
		return target
	}
	for packagePath, StructInfo := range target {
		if _, ok := source[packagePath]; ok {
			for structName, structFields := range StructInfo {
				if _, ok := source[structName]; !ok {
					source[packagePath][structName] = structFields
				}
			}
		} else {
			source[packagePath] = StructInfo
		}
	}
	return source
}

func import_merge(source, target map[string]string) map[string]string {
	if source == nil || len(source) == 0 {
		return target
	}
	for key, val := range target {
		if _, ok := source[key]; !ok {
			source[key] = val
		}
	}
	return source
}
