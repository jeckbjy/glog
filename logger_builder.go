package glog

import (
	"context"
	"fmt"
	"sync"
)

var gBuilderPool = sync.Pool{
	New: func() interface{} {
		return &Builder{}
	},
}

func newBuilder(e *Entry) *Builder {
	b := gBuilderPool.New().(*Builder)
	b.entry = e
	return b
}

// Builder 可通过With等方法构建Logger
type Builder struct {
	entry *Entry
}

func (b *Builder) free() {
	gBuilderPool.Put(b)
}

func (b *Builder) With(fields ...Field) *Builder {
	b.entry.Fields = append(b.entry.Fields, fields...)
	return b
}

func (b *Builder) WithCallDepth(depth int) *Builder {
	b.entry.CallDepth = depth
	return b
}

func (b *Builder) Log(lv Level, args ...interface{}) {
	e := b.entry
	l := e.Logger
	if l.IsEnable(lv) {
		text := fmt.Sprint(args...)
		e := b.entry
		e.Level = lv
		e.Text = text
		l.Write(e)
	}
	b.free()
}

func (b *Builder) Logf(lv Level, format string, args ...interface{}) {
	e := b.entry
	l := e.Logger

	if l.IsEnable(lv) {
		text := fmt.Sprintf(format, args...)
		e.Level = lv
		e.Text = text
		l.Write(e)
	}
	b.free()
}

func (b *Builder) Logw(lv Level, msg string, args ...interface{}) {
	e := b.entry
	l := e.Logger
	if l.IsEnable(lv) {
		e.Level = lv
		e.Text = msg
		if len(args) > 0 {
			e.Fields = append(e.Fields, toFields(args)...)
		}
		l.Write(e)
	}
	b.free()
}

func (b *Builder) Trace(ctx context.Context, args ...interface{}) {
	b.Log(TraceLevel, args...)
}

func (b *Builder) Debug(ctx context.Context, args ...interface{}) {
	b.Log(DebugLevel, args...)
}

func (b *Builder) Info(ctx context.Context, args ...interface{}) {
	b.Log(InfoLevel, args...)
}

func (b *Builder) Warn(ctx context.Context, args ...interface{}) {
	b.Log(WarnLevel, args...)
}

func (b *Builder) Error(ctx context.Context, args ...interface{}) {
	b.Log(ErrorLevel, args...)
}

func (b *Builder) Fatal(ctx context.Context, args ...interface{}) {
	b.Log(FatalLevel, args...)
}

func (b *Builder) Tracef(ctx context.Context, format string, args ...interface{}) {
	b.Logf(TraceLevel, format, args...)
}

func (b *Builder) Debugf(format string, args ...interface{}) {
	b.Logf(DebugLevel, format, args...)
}

func (b *Builder) Infof(ctx context.Context, format string, args ...interface{}) {
	b.Logf(InfoLevel, format, args...)
}

func (b *Builder) Warnf(ctx context.Context, format string, args ...interface{}) {
	b.Logf(WarnLevel, format, args...)
}

func (b *Builder) Errorf(ctx context.Context, format string, args ...interface{}) {
	b.Logf(ErrorLevel, format, args...)
}

func (b *Builder) Fatalf(ctx context.Context, format string, args ...interface{}) {
	b.Logf(FatalLevel, format, args...)
}

func (b *Builder) Tracew(ctx context.Context, format string, args ...interface{}) {
	b.Logf(TraceLevel, format, args...)
}

func (b *Builder) Debugw(format string, args ...interface{}) {
	b.Logf(DebugLevel, format, args...)
}

func (b *Builder) Infow(ctx context.Context, format string, args ...interface{}) {
	b.Logf(InfoLevel, format, args...)
}

func (b *Builder) Warnw(ctx context.Context, format string, args ...interface{}) {
	b.Logf(WarnLevel, format, args...)
}

func (b *Builder) Errorw(ctx context.Context, format string, args ...interface{}) {
	b.Logf(ErrorLevel, format, args...)
}

func (b *Builder) Fatalw(ctx context.Context, format string, args ...interface{}) {
	b.Logf(FatalLevel, format, args...)
}
