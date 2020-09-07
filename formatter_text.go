package glog

const defaultTextLayout = "%-5p %d{yyyy-MM-dd HH:mm:ss} %F:%L %x{*}%w - %m%n"

// NewTextFormatter 通过Layout创建Formatter
func NewTextFormatter(layout string) (Formatter, error) {
	l, err := NewLayout(layout)
	if err != nil {
		return nil, err
	}
	f := &textFormatter{layout: l}
	return f, nil
}

// MustNewTextFormatter 通过Layout创建Formatter,若失败则抛出异常
func MustNewTextFormatter(layout string) Formatter {
	f, err := NewTextFormatter(layout)
	if err != nil {
		panic(err)
	}

	return f
}

type textFormatter struct {
	layout *Layout
}

func (f *textFormatter) Name() string {
	return "text"
}

func (f *textFormatter) Format(e *Entry) ([]byte, error) {
	data := f.layout.Format(e)
	return data, nil
}
