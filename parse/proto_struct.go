package parse

import (
	"fmt"
	"reflect"
	"strings"

	_struct "github.com/golang/protobuf/ptypes/struct"
)

func DecodeProtoStruct2Map(protoStruct *_struct.Struct) map[string]interface{} {
	if protoStruct == nil {
		return nil
	}
	mapParam := map[string]interface{}{}
	for key, val := range protoStruct.Fields {
		mapParam[key] = DecodeProtoValue2Interface(val)
	}
	return mapParam
}

func DecodeProtoValue2Interface(protoStruct *_struct.Value) interface{} {
	if protoStruct == nil {
		return nil
	}
	switch kind := protoStruct.Kind.(type) {
	case *_struct.Value_NullValue:
		return nil
	case *_struct.Value_NumberValue:
		return kind.NumberValue
	case *_struct.Value_StringValue:
		return kind.StringValue
	case *_struct.Value_BoolValue:
		return kind.BoolValue
	case *_struct.Value_StructValue:
		return DecodeProtoStruct2Map(kind.StructValue)
	case *_struct.Value_ListValue:
		Interface := make([]interface{}, len(kind.ListValue.Values))
		for key, val := range kind.ListValue.Values {
			Interface[key] = DecodeProtoValue2Interface(val)
		}
		return Interface
	default:
		panic("protos_truct: unknown kind")
	}
}

// EncodeMap2ProtoStruct converts a map[string]interface{} to a ptypes.Struct
func EncodeMap2ProtoStruct(mapParam map[string]interface{}) *_struct.Struct {
	if mapParam == nil {
		return nil
	}
	fields := make(map[string]*_struct.Value, len(mapParam))
	for key, val := range mapParam {
		fields[key] = EncodeInterface2ProtoValue(val)
	}
	return &_struct.Struct{
		Fields: fields,
	}
}

// EncodeInterface2ProtoStruct converts an interface{} to a ptypes.Value
func EncodeInterface2ProtoValue(interfaceParam interface{}) *_struct.Value {
	if interfaceParam == nil {
		return nil
	}
	switch value := interfaceParam.(type) {
	case bool:
		return &_struct.Value{
			Kind: &_struct.Value_BoolValue{
				BoolValue: value,
			},
		}
	case error:
		return &_struct.Value{
			Kind: &_struct.Value_StringValue{
				StringValue: value.Error(),
			},
		}
	case string:
		return &_struct.Value{
			Kind: &_struct.Value_StringValue{
				StringValue: value,
			},
		}
	default:
		return toValue(reflect.ValueOf(value))
	}
}

func toValue(value reflect.Value) *_struct.Value {
	switch value.Kind() {
	case reflect.Int, reflect.Int32, reflect.Int64:
		return &_struct.Value{
			Kind: &_struct.Value_NumberValue{
				NumberValue: float64(value.Int()),
			},
		}
	case reflect.Uint, reflect.Uint32, reflect.Uint64:
		return &_struct.Value{
			Kind: &_struct.Value_NumberValue{
				NumberValue: float64(value.Uint()),
			},
		}
	case reflect.Float32, reflect.Float64:
		return &_struct.Value{
			Kind: &_struct.Value_NumberValue{
				NumberValue: value.Float(),
			},
		}
	case reflect.Array, reflect.Slice:
		if value.IsNil() {
			return nil
		}
		values := make([]*_struct.Value, value.Len())
		for index := range values {
			values[index] = toValue(value.Index(index))
		}
		return &_struct.Value{
			Kind: &_struct.Value_ListValue{
				ListValue: &_struct.ListValue{
					Values: values,
				},
			},
		}
	case reflect.Map:
		if value.IsNil() {
			return nil
		}
		fields := make(map[string]*_struct.Value, len(value.MapKeys()))
		for _, key := range value.MapKeys() {
			if key.Kind() == reflect.String {
				fields[key.String()] = toValue(value.MapIndex(key))
			}
		}
		return &_struct.Value{
			Kind: &_struct.Value_StructValue{
				StructValue: &_struct.Struct{
					Fields: fields,
				},
			},
		}
	case reflect.Struct:
		t := value.Type()
		size := value.NumField()
		if size == 0 {
			return nil
		}
		fields := make(map[string]*_struct.Value, size)
		for i := 0; i < size; i++ {
			field := t.Field(i)
			// 支持内嵌结构体展开
			if field.Anonymous && (field.Type.Kind() == reflect.Struct || field.Type.Kind() == reflect.Ptr) {
				if _, ok := field.Tag.Lookup("protoOpen"); ok {

					sizeF := field.Type.NumField()
					typeF := field.Type

					for j := 0; j < sizeF; j++ {
						tagName := findTagNameJsonOrDefault(typeF.Field(j))
						fields[tagName] = toValue(value.FieldByIndex([]int{i, j}))
					}
					continue
				}
			}
			name := field.Name
			tagName := findTagNameJsonOrDefault(field)
			if len(name) > 0 && 'A' <= name[0] && name[0] <= 'Z' {
				fields[tagName] = toValue(value.Field(i))
			}
		}
		return &_struct.Value{
			Kind: &_struct.Value_StructValue{
				StructValue: &_struct.Struct{
					Fields: fields,
				},
			},
		}
	case reflect.Ptr:
		if value.IsNil() {
			return nil
		}
		return toValue(reflect.Indirect(value))
	case reflect.Interface:
		if value.IsNil() {
			return nil
		}
		return toValue(value.Elem())
	default:
		return &_struct.Value{
			Kind: &_struct.Value_StringValue{
				StringValue: fmt.Sprint(value),
			},
		}
	}
}

func findTagNameJsonOrDefault(f reflect.StructField) string {
	defaultName := f.Name

	tagName := f.Tag.Get("json")
	if len(tagName) == 0 {
		return defaultName
	}

	tagName = strings.Split(tagName, ",")[0]
	if tagName == "-" {
		tagName = defaultName
	}

	return tagName
}
