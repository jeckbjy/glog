package glog

import "context"

var defaultLogger = NewDefault()

// SetDefault 设置默认的log
func SetDefault(l Logger) {
	if defaultLogger != nil {
		defaultLogger.Stop()
	}
	defaultLogger = l
}

// NewDefault 创建默认的Logger,默认只包含Console的输出通路
func NewDefault() Logger {
	conf := NewConfig()
	conf.Channels = append(conf.Channels, NewConsoleChannel())
	return NewLogger(conf)
}

func With(fields ...Field) *Builder {
	e := NewEntry(defaultLogger)
	b := newBuilder(e)
	return b
}

func Trace(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.Log(ctx, TraceLevel, msg, fields...)
}

func Debug(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.Log(ctx, DebugLevel, msg, fields...)
}

func Info(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.Log(ctx, InfoLevel, msg, fields...)
}

func Warn(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.Log(ctx, WarnLevel, msg, fields...)
}

func Error(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.Log(ctx, ErrorLevel, msg, fields...)
}

func Fatal(ctx context.Context, msg string, fields ...Field) {
	defaultLogger.Log(ctx, FatalLevel, msg, fields...)
}

func Tracef(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Logf(ctx, TraceLevel, format, args...)
}

func Debugf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Logf(ctx, DebugLevel, format, args...)
}

func Infof(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Logf(ctx, InfoLevel, format, args...)
}

func Warnf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Logf(ctx, WarnLevel, format, args...)
}

func Errorf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Logf(ctx, ErrorLevel, format, args...)
}

func Fatalf(ctx context.Context, format string, args ...interface{}) {
	defaultLogger.Logf(ctx, FatalLevel, format, args...)
}

func Tracew(ctx context.Context, msg string, args ...interface{}) {
	defaultLogger.Logw(ctx, TraceLevel, msg, args...)
}

func Debugw(ctx context.Context, msg string, args ...interface{}) {
	defaultLogger.Logw(ctx, TraceLevel, msg, args...)
}

func Infow(ctx context.Context, msg string, args ...interface{}) {
	defaultLogger.Logw(ctx, InfoLevel, msg, args...)
}

func Warnw(ctx context.Context, msg string, args ...interface{}) {
	defaultLogger.Logw(ctx, WarnLevel, msg, args...)
}

func Errorw(ctx context.Context, msg string, args ...interface{}) {
	defaultLogger.Logw(ctx, ErrorLevel, msg, args...)
}

func Fatalw(ctx context.Context, msg string, args ...interface{}) {
	defaultLogger.Logw(ctx, FatalLevel, msg, args...)
}
