package glog

import (
	"fmt"
	"strings"
)

const (
	actionFields = 'w'
	actionTags   = 'x'
)

// NewLayout 创建Layout
func NewLayout(format string) (*Layout, error) {
	l := &Layout{}
	if err := l.Parse(format); err != nil {
		return nil, err
	}

	return l, nil
}

// https://wiki.jikexueyuan.com/project/log4j/log4j-patternlayout.html
// 实现类似log4j的PatternLayout格式,比如 %r [%t] %p %c %x - %m%n
// TODO: 支持配色
// https://blog.csdn.net/qq_40147863/article/details/88880053
type Layout struct {
	actions []*Action
}

// Action {$prefix}%x{-$min.$max}token{$param}
type Action struct {
	Key    byte        // %xx
	Prefix string      // xx%
	Min    int         // 最小宽度
	Max    int         // 最大宽度
	Param  string      // %x{param},原始参数
	Data   interface{} // 计算处理后数据
}

func (l *Layout) Parse(format string) error {
	actions := make([]*Action, 0, 8)
	lex := lexer{}
	lex.init(format)
	for !lex.eof() {
		act := &Action{}
		act.Prefix = lex.readPrefix()
		if lex.eof() {
			if act.Prefix != "" {
				actions = append(actions, act)
			}
			break
		}

		act.Min = lex.readNumber()
		if lex.readExpect('.') {
			act.Max = lex.readNumber()
		}
		act.Key = lex.readKey()
		if lex.readExpect('{') {
			if p, err := lex.readCloseTerm('}'); err != nil {
				return err
			} else {
				act.Param = p
			}
		}
		if act.Max < 0 {
			return fmt.Errorf("invalid max, %+d", act.Max)
		}
		if act.Key == 0 {
			return fmt.Errorf("invalid key")
		}
		if act.Key == 'd' {
			if df, err := NewDateFormat(act.Param); err != nil {
				return err
			} else {
				act.Data = df
			}
		}
		actions = append(actions, act)
	}

	l.actions = actions
	return nil
}

func (l *Layout) Format(e *Entry) []byte {
	buf := NewBuffer()
	for _, a := range l.actions {
		if a.Prefix != "" {
			buf.AppendString(a.Prefix)
		}

		switch a.Key {
		case '%':
			buf.AppendByte('%')
		case 'n':
			buf.AppendByte('\n')
		case 'm':
			buf.Put(a.Min, a.Max, e.Text)
		case 'p':
			buf.Put(a.Min, a.Max, e.Level.String())
		case 'F':
			buf.Put(a.Min, a.Max, e.File)
		case 'L':
			buf.Put(a.Min, a.Max, toString(e.Line))
		case 'M':
			buf.Put(a.Min, a.Max, e.Method)
		case 'l':
			buf.Putf(a.Min, a.Max, "%s(%s:%d)", e.Method, e.File, e.Line)
		case 'd':
			date := a.Data.(*DateFormat).Format(e.Time)
			buf.Put(a.Min, a.Max, date)
		case 'x':
			if a.Param == "*" {
				for i := 0; i < e.Tags.Len(); i++ {
					if i > 0 {
						buf.AppendByte(' ')
					}
					buf.AppendString(e.Tags.GetKey(i))
					buf.AppendByte('=')
					buf.AppendString(e.Tags.GetValue(i))
				}
			} else if a.Param != "" {
				if value, ok := e.Tags.Get(a.Param); ok {
					buf.AppendString(value)
				}
			}
		case 'w':
			// TODO:通过参数控制分隔符
			if len(e.Fields) > 0 {
				for i, v := range e.Fields {
					if i > 0 {
						buf.AppendByte(' ')
					}
					buf.AppendString(v.Key)
					buf.AppendByte('=')
					v.AppendValueToBuffer(buf)
				}
			}
		}
	}

	return buf.Bytes()
}

type lexer struct {
	format string
	cur    int
	end    int
}

func (l *lexer) init(f string) {
	l.format = f
	l.cur = 0
	l.end = len(f)
}

func (l *lexer) eof() bool {
	return l.cur >= l.end
}

func (l *lexer) readExpect(ch byte) bool {
	if l.cur < l.end && l.format[l.cur] == ch {
		l.cur++
		return true
	}

	return false
}

func (l *lexer) readPrefix() string {
	beg := l.cur
	for ; l.cur < l.end; l.cur++ {
		if l.format[l.cur] == '%' {
			res := ""
			if beg < l.cur {
				res = l.format[beg:l.cur]
			}

			l.cur++

			return res
		}
	}

	l.cur = l.end
	return l.format[beg:]
}

func (l *lexer) readNumber() int {
	res := 0
	neg := false
	cur := l.cur
	switch ch := l.format[cur]; {
	case ch == '-':
		neg = true
		cur++
	case isDigit(ch):
		res = toDigit(ch)
		cur++
	default:
		return 0
	}

	for ; cur < l.end; cur++ {
		if !isDigit(l.format[cur]) {
			break
		}
		res = res*10 + toDigit(l.format[cur])
	}

	if neg {
		res = -res
	}

	l.cur = cur
	return res
}

func (l *lexer) readKey() byte {
	if l.cur < l.end {
		ch := l.format[l.cur]
		l.cur++
		return ch
	}

	return 0
}

// readCloseTerm 读取到结束符
func (l *lexer) readCloseTerm(ch byte) (string, error) {
	for cur := l.cur; cur < l.end; cur++ {
		if l.format[cur] == ch {
			res := strings.TrimSpace(l.format[l.cur:cur])
			l.cur = cur + 1
			return res, nil
		}
	}

	return "", fmt.Errorf("not find close term, offset=%+v", l.cur)
}
