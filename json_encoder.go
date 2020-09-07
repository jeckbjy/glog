package glog

import (
	"fmt"
	"math"
	"unicode/utf8"
)

// For JSON-escaping; see jsonEncoder.safeAddString below.
const _hex = "0123456789abcdef"

// NewJsonEncoder 创建json编码器
func NewJsonEncoder() JsonEncoder {
	return JsonEncoder{buf: NewBuffer(), spaced: false}
}

// JsonEncoder 简单的Json编码器,通常只需要一级
// TODO:支持多级
type JsonEncoder struct {
	buf    *Buffer
	spaced bool
}

func (enc *JsonEncoder) Bytes() []byte {
	return enc.buf.Bytes()
}

func (enc *JsonEncoder) AddField(f *Field, fn func(string) string) {
	key := f.Key
	if fn != nil {
		key = fn(key)
	}
	switch f.Type {
	case FieldTypeString:
		enc.AddString(key, f.String)
	case FieldTypeByte:
		enc.AddString(key, string(byte(f.Int)))
	case FieldTypeBool:
		enc.AddBool(key, f.Int == 1)
	case FieldTypeInt:
		enc.AddInt(key, f.Int)
	case FieldTypeUint:
		enc.AddUint(key, uint64(f.Int))
	case FieldTypeFloat32:
		enc.AddFloat32(key, math.Float32frombits(uint32(f.Int)))
	case FieldTypeFloat64:
		enc.AddFloat64(key, math.Float64frombits(uint64(f.Int)))
	case FieldTypeAny:
		switch v := f.Value.(type) {
		case string:
			enc.AddString(key, v)
		case bool:
			enc.AddBool(key, v)
		case byte:
			enc.AddString(key, string(v))
		case int:
			enc.AddInt(key, int64(v))
		case int8:
			enc.AddInt(key, int64(v))
		case int16:
			enc.AddInt(key, int64(v))
		case int32:
			enc.AddInt(key, int64(v))
		case int64:
			enc.AddInt(key, v)
		case uint:
			enc.AddUint(key, uint64(v))
		case uint16:
			enc.AddUint(key, uint64(v))
		case uint32:
			enc.AddUint(key, uint64(v))
		case uint64:
			enc.AddUint(key, uint64(v))
		case float32:
			enc.AddFloat32(key, v)
		case float64:
			enc.AddFloat64(key, v)
		default:
			// 只支持普通类型,复杂结构当string处理?
			enc.AddString(key, fmt.Sprintf("%+v", v))
		}
	}
}

func (enc *JsonEncoder) Begin() {
	enc.buf.AppendByte('{')
}

func (enc *JsonEncoder) End() {
	enc.buf.AppendByte('}')
}

// AddValidString 添加非空字符串
func (enc *JsonEncoder) AddValidString(key string, val string) {
	if len(val) > 0 {
		enc.AddString(key, val)
	}
}

func (enc *JsonEncoder) AddString(key string, val string) {
	enc.addKey(key)
	enc.buf.AppendByte('"')
	enc.safeAddString(val)
	enc.buf.AppendByte('"')
}

func (enc *JsonEncoder) AddBool(key string, val bool) {
	enc.addKey(key)
	enc.buf.AppendBool(val)
}

func (enc *JsonEncoder) AddInt(key string, val int64) {
	enc.addKey(key)
	enc.buf.AppendInt(val)
}

func (enc *JsonEncoder) AddUint(key string, val uint64) {
	enc.addKey(key)
	enc.buf.AppendUint(val)
}

func (enc *JsonEncoder) AddFloat32(key string, val float32) {
	enc.addKey(key)
	enc.buf.AppendFloat32(val)
}

func (enc *JsonEncoder) AddFloat64(key string, val float64) {
	enc.addKey(key)
	enc.buf.AppendFloat64(val)
}

func (enc *JsonEncoder) AddComplex128(key string, val complex128) {
	enc.addKey(key)
	enc.buf.AppendByte('"')
	enc.buf.AppendComplex128(val)
	enc.buf.AppendByte('"')
}

func (enc *JsonEncoder) addKey(key string) {
	enc.addElementSeparator()
	enc.buf.AppendByte('"')
	enc.safeAddString(key)
	enc.buf.AppendByte('"')
	enc.buf.AppendByte(':')
	if enc.spaced {
		enc.buf.AppendByte(' ')
	}
}

func (enc *JsonEncoder) addElementSeparator() {
	last := enc.buf.Last()
	if last == 0 {
		return
	}
	switch last {
	case '{', '[', ':', ',', ' ':
		return
	default:
		enc.buf.AppendByte(',')
		if enc.spaced {
			enc.buf.AppendByte(' ')
		}
	}
}

// safeAddString JSON-escapes a string and appends it to the internal buffer.
// Unlike the standard library's encoder, it doesn't attempt to protect the
// user from browser vulnerabilities or JSONP-related problems.
func (enc *JsonEncoder) safeAddString(s string) {
	for i := 0; i < len(s); {
		if enc.tryAddRuneSelf(s[i]) {
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if enc.tryAddRuneError(r, size) {
			i++
			continue
		}
		enc.buf.AppendString(s[i : i+size])
		i += size
	}
}

// tryAddRuneSelf appends b if it is valid UTF-8 character represented in a single byte.
func (enc *JsonEncoder) tryAddRuneSelf(b byte) bool {
	if b >= utf8.RuneSelf {
		return false
	}
	if 0x20 <= b && b != '\\' && b != '"' {
		enc.buf.AppendByte(b)
		return true
	}
	switch b {
	case '\\', '"':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte(b)
	case '\n':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('n')
	case '\r':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('r')
	case '\t':
		enc.buf.AppendByte('\\')
		enc.buf.AppendByte('t')
	default:
		// Encode bytes < 0x20, except for the escape sequences above.
		enc.buf.AppendString(`\u00`)
		enc.buf.AppendByte(_hex[b>>4])
		enc.buf.AppendByte(_hex[b&0xF])
	}
	return true
}

func (enc *JsonEncoder) tryAddRuneError(r rune, size int) bool {
	if r == utf8.RuneError && size == 1 {
		enc.buf.AppendString(`\ufffd`)
		return true
	}
	return false
}
