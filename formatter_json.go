package glog

import (
	"encoding/csv"
	"fmt"
	"strings"
)

// NewJsonForamtter 通过Layout创建Json Formatter
func NewJsonFormatter(layout string) (Formatter, error) {
	f := &jsonFormatter{}
	if err := f.Parse(layout); err != nil {
		return nil, err
	}

	return f, nil
}

// MustNewJsonFormatter ...
func MustNewJsonFormatter(layout string) Formatter {
	f, err := NewTextFormatter(layout)
	if err != nil {
		panic(err)
	}

	return f
}

type jsonField struct {
	Key    string
	Format *Layout
}

// jsonFormatter json编码,配置通过Layout解析获得,要求是key=value结构
// Layout格式例子: msg=%m level=%p file=%F
// Tags和Fields都会展开输出
type jsonFormatter struct {
	fields []jsonField
}

func (f *jsonFormatter) Name() string {
	return "json"
}

func (f *jsonFormatter) Format(e *Entry) ([]byte, error) {
	enc := NewJsonEncoder()
	enc.Begin()
	for _, f := range f.fields {
		// 都以字符串的形式输出
		value := f.Format.Format(e)
		if len(value) > 0 {
			enc.AddString(f.Key, string(value))
		}
	}

	if !e.Tags.Empty() {
		size := e.Tags.Len()
		for i := 0; i < size; i++ {
			key, value := e.Tags.GetAt(i)
			enc.AddString(key, value)
		}
	}

	if len(e.Fields) > 0 {
		for _, f := range e.Fields {
			enc.AddField(&f, nil)
		}
	}

	enc.End()
	return enc.Bytes(), nil
}

func (f *jsonFormatter) Parse(layout string) error {
	// time="%d" text=%t
	r := csv.NewReader(strings.NewReader(layout))
	r.Comma = ' '
	fields, err := r.Read()
	if err != nil {
		return err
	}

	var key, value string
	for _, field := range fields {
		tokens := strings.SplitN(field, "=", 2)
		switch len(tokens) {
		case 1:
			key = ""
			value = strings.TrimSpace(tokens[0])
		case 2:
			key = strings.TrimSpace(tokens[0])
			value = strings.TrimSpace(tokens[1])
		default:
			return fmt.Errorf("invalid layout")
		}

		if len(value) < 2 || value[0] != '%' {
			return fmt.Errorf("invalid layout")
		}

		// 忽略Tags和Fields,自动全部展开
		if value[1] == actionFields || value[1] == actionTags {
			continue
		}

		l, err1 := NewLayout(value)
		if err1 != nil {
			return err1
		}

		f.fields = append(f.fields, jsonField{Key: key, Format: l})
	}

	return nil
}
