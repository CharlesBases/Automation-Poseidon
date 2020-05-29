package utils

import (
	"strings"
)

// aaa_bbb_ccc to AaaBbbCcc
func Title(name string) string {
	builder := strings.Builder{}
	for _, val := range strings.Split(name, "_") {
		builder.WriteString(strings.Title(val))
	}
	return builder.String()
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

func parsefield(gotype string) (prefix, structname string) {
	gotype = strings.TrimPrefix(strings.TrimPrefix(strings.TrimPrefix(gotype, "*"), "[]"), "*")
	if index := strings.LastIndex(gotype, "."); index != -1 {
		return gotype[:index], gotype[index+1:]
	}
	return "", gotype
}
