package glog

import (
	"fmt"
	"os"
	"runtime"
)

// Foreground colors.
const (
	Black Color = iota + 30
	Red
	Green
	Yellow
	Blue
	Magenta // 品红
	Cyan    // 青色
	White
)

// Color represents a text color.
type Color uint8

// NewConsoleChannel 创建控制台输出Channel
func NewConsoleChannel(opts ...ChannelOption) Channel {
	o := NewChannelOptions(opts...)
	c := &consoleChannel{writer: os.Stdout}
	c.color = runtime.GOOS != "windows" && IsTerminal(int(c.writer.Fd()))
	c.Init(o)
	return c
}

var levelToColor = map[Level]Color{
	TraceLevel: Blue,
	DebugLevel: Blue,
	InfoLevel:  Cyan,
	WarnLevel:  Yellow,
	ErrorLevel: Red,
	FatalLevel: Magenta,
}

// consoleChannel 控制台输出
type consoleChannel struct {
	BaseChannel
	writer *os.File
	color  bool
}

func (c *consoleChannel) Name() string {
	return "console"
}

func (c *consoleChannel) Write(e *Entry) {
	if text := c.Format(e); len(text) != 0 {
		// auto change line
		if text[len(text)-1] != '\n' {
			text = append(text, '\n')
		}

		if c.color {
			text = c.AddColor(e.Level, text)
		}

		c.writer.Write(text)
	}
}

func (c *consoleChannel) AddColor(lv Level, text []byte) []byte {
	result := fmt.Sprintf("\x1b[%dm%s\x1b[0m", uint8(levelToColor[lv]), text)
	return []byte(result)
}
