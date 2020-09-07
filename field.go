package glog

import "math"

type FieldType uint8

const (
	FieldTypeUnknown FieldType = iota
	FieldTypeAny
	FieldTypeString
	FieldTypeByte
	FieldTypeBool
	FieldTypeInt
	FieldTypeUint
	FieldTypeFloat64
	FieldTypeFloat32
)

type Field struct {
	Key    string
	Type   FieldType
	Int    int64
	String string
	Value  interface{}
}

func (f *Field) AppendValueToBuffer(b *Buffer) {
	switch f.Type {
	case FieldTypeString:
		b.AppendString(f.String)
	case FieldTypeByte:
		b.AppendByte(byte(f.Int))
	case FieldTypeBool:
		b.AppendBool(f.Int == 1)
	case FieldTypeInt:
		b.AppendInt(f.Int)
	case FieldTypeUint:
		b.AppendUint(uint64(f.Int))
	case FieldTypeFloat32:
		b.AppendFloat32(math.Float32frombits(uint32(f.Int)))
	case FieldTypeFloat64:
		b.AppendFloat64(math.Float64frombits(uint64(f.Int)))
	case FieldTypeAny:
		switch v := f.Value.(type) {
		case string:
			b.AppendString(v)
		case bool:
			b.AppendBool(v)
		case byte:
			b.AppendByte(v)
		case int:
			b.AppendInt(int64(v))
		case int8:
			b.AppendInt(int64(v))
		case int16:
			b.AppendInt(int64(v))
		case int32:
			b.AppendInt(int64(v))
		case int64:
			b.AppendInt(v)
		case uint:
			b.AppendUint(uint64(v))
		case uint16:
			b.AppendUint(uint64(v))
		case uint32:
			b.AppendUint(uint64(v))
		case uint64:
			b.AppendUint(uint64(v))
		case float32:
			b.AppendFloat32(v)
		case float64:
			b.AppendFloat64(v)
		default:
			b.Appendf("%+v", v)
		}
	}
}

func Any(key string, value interface{}) Field {
	return Field{Key: key, Type: FieldTypeAny, Value: value}
}

func String(key string, val string) Field {
	return Field{Key: key, Type: FieldTypeString, String: val}
}

func Byte(key string, val byte) Field {
	return Field{Key: key, Type: FieldTypeByte, Int: int64(val)}
}

func Bool(key string, val bool) Field {
	var i int64
	if val {
		i = 1
	}
	return Field{Key: key, Type: FieldTypeBool, Int: i}
}

func Int(key string, val int) Field {
	return Field{Key: key, Type: FieldTypeInt, Int: int64(val)}
}

func Int8(key string, val int8) Field {
	return Field{Key: key, Type: FieldTypeInt, Int: int64(val)}
}

func Int16(key string, val int16) Field {
	return Field{Key: key, Type: FieldTypeInt, Int: int64(val)}
}

func Int32(key string, val int32) Field {
	return Field{Key: key, Type: FieldTypeInt, Int: int64(val)}
}

func Int64(key string, val int64) Field {
	return Field{Key: key, Type: FieldTypeInt, Int: int64(val)}
}

func Uint(key string, val uint) Field {
	return Field{Key: key, Type: FieldTypeUint, Int: int64(val)}
}

func Uint8(key string, val uint8) Field {
	return Field{Key: key, Type: FieldTypeUint, Int: int64(val)}
}

func Uint16(key string, val uint16) Field {
	return Field{Key: key, Type: FieldTypeUint, Int: int64(val)}
}

func Uint32(key string, val uint32) Field {
	return Field{Key: key, Type: FieldTypeUint, Int: int64(val)}
}

func Uint64(key string, val uint64) Field {
	return Field{Key: key, Type: FieldTypeUint, Int: int64(val)}
}

func Float32(key string, val float32) Field {
	return Field{Key: key, Type: FieldTypeFloat32, Int: int64(math.Float32bits(val))}
}

func Float64(key string, val float64) Field {
	return Field{Key: key, Type: FieldTypeFloat64, Int: int64(math.Float64bits(val))}
}

// toFields 若args[i]为Field则直接添加,否则认为是Key=Value的形式,如果不匹配,则会统一放到错误字段中
func toFields(args []interface{}) []Field {
	if len(args) == 0 {
		return nil
	}

	invalid := NewBuffer()

	fields := make([]Field, 0, len(args))
	for i := 0; i < len(args); {
		if f, ok := args[i].(Field); ok {
			fields = append(fields, f)
			i++
			continue
		}
		if i == len(args)-1 {
			if !invalid.Empty() {
				invalid.AppendByte(',')
			}
			invalid.Appendf("without value[%d]", i)
			break
		}
		key, val := args[i], args[i+1]
		if keystr, ok := key.(string); ok {
			fields = append(fields, Any(keystr, val))
		} else {
			if !invalid.Empty() {
				invalid.AppendByte(',')
			}
		}
		i += 2
	}
	if !invalid.Empty() {
		fields = append(fields, Field{})
	}

	return fields
}
