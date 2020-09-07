package glog

import (
	"context"
	"fmt"
	"path/filepath"
	"sync"
	"sync/atomic"
	"time"
)

var (
	DefaultFormatter = MustNewTextFormatter(defaultTextLayout)
)

const (
	DefaultMax       = 10000 // 默认日志最大条数
	DefaultCallDepth = 3     // 堆栈深度,忽略log相关的堆栈
)

// Entry 一条输出日志
type Entry struct {
	sync.RWMutex
	Logger    Logger               // 日志Owner
	Level     Level                // 日志级别
	Text      string               // 日志信息
	Tags      SortedMap            // Tags,初始化时设置的标签信息,比如env,host等
	Fields    []Field              // 附加字段,无序,k=v格式整体输出
	Time      time.Time            // 时间戳
	Context   context.Context      // 上下文,通常用于填充Fields
	Host      string               // 配置host
	Path      string               // 文件全路径,包含文件名
	File      string               // 文件名
	Line      int                  // 行号
	Method    string               // 方法名
	CallDepth int                  // 需要忽略的堆栈
	outputs   map[Formatter][]byte // 相同的Formater只会构建一次
	refs      int32                // 引用计数,当为0时,会放到缓存中
}

var gEntryPool = sync.Pool{
	New: func() interface{} {
		return &Entry{}
	},
}

// NewEntry 创建Entry
func NewEntry(logger Logger) *Entry {
	e := gEntryPool.Get().(*Entry)
	e.Logger = logger
	e.Time = time.Now()
	e.CallDepth = DefaultCallDepth
	e.outputs = make(map[Formatter][]byte)
	e.refs = 1
	return e
}

// Obtain 增加引用计数
func (e *Entry) Obtain() {
	atomic.AddInt32(&e.refs, 1)
}

// Free 当引用计数为0
func (e *Entry) Free() {
	if atomic.AddInt32(&e.refs, -1) <= 0 {
		gEntryPool.Put(e)
	}
}

// Channel 代表日志输出通路
type Channel interface {
	IsEnable(lv Level) bool
	Level() Level
	SetLevel(lv Level)
	Name() string
	Open() error
	Close() error
	Write(msg *Entry)
}

// BatchChannel 以batch形式发送,有需要的可以
type BatchChannel interface {
	Channel
	WriteBatch(msg []*Entry)
}

// Filter 在每天日志写入Channel前统一预处理,若返回错误则忽略该条日志
// 可用于通过Context添加Field,对某些Field加密等处理
type Filter func(*Entry) error

// Formatter 用于格式化Entry
// 相同的Formatter对每个Entry只会格式化一次
type Formatter interface {
	Name() string                      //
	Format(msg *Entry) ([]byte, error) // 格式化输出
}

// Sampler 日志采样,对于高频的日志可以限制发送频率
type Sampler interface {
	Check(msg *Entry) bool
}

// Config 配置信息
type Config struct {
	Channels      []Channel // 日志输出通路,至少1个,默认Console
	Filters       []Filter  // 过滤函数
	Tags          SortedMap // 全局Fields,比如env,cluster,psm,host等
	Level         Level     // 日志级别,默认Info
	LogMax        int       // 最大缓存日志数
	DisableCaller bool      // 是否关闭Caller,若为true则获取不到文件名等信息
	Async         bool      // 是否异步,默认同步
}

func (c *Config) AddChannels(channels ...Channel) {
	c.Channels = append(c.Channels, channels...)
}

// AddTags 添加Tags
func (c *Config) AddTags(tags map[string]string) {
	c.Tags.Fill(tags)
}

// NewConfig 创建配置
func NewConfig() *Config {
	return &Config{
		Level:  TraceLevel,
		LogMax: DefaultMax,
	}
}

// Logger 日志系统接口(structured, leveled logging)
// 1:支持同步模式和异步模式
// 	在Debug模式下,通常使用同步模式,因为可以保证console与fmt顺序一致
// 	正式环境下可以使用异步模式,保证日志不会影响服务质量,不能保证日志不丢失
// 2:配置信息
//	大部分配置信息是不能动态更新的,异步处理时并没有枷锁,比如Channel,Filter,Tags
//	部分简单配置是可以动态更新的,比如Level降级,可用于临时调试
// 3:关于Context
//	通过Context可以透传RequestID,LogID，UID等信息,Log第一个参数都强制要求传入Ctx,但可以为nil
// 	Logger本身并不知道如何处理Ctx,因此需要初始化时手动添加Filter用来解析Context
// 4:类似zap,接口上提供了三套接口,Log,Logf,Logw
//  Log要求显示结构化日志输出,而不是Log(ctx context.Context,args ...interface{})这样的形式
//	Logf与普通日志系统类似
//	Logw会自动将Filed或者i,i+1转换为key=value形式,要求key是字符串类型
type Logger interface {
	IsEnable(lv Level) bool
	SetLevel(name string, lv Level)
	Start()
	Stop()
	Write(e *Entry)
	Log(ctx context.Context, lv Level, msg string, fields ...Field)
	Logf(ctx context.Context, lv Level, format string, args ...interface{})
	Logw(ctx context.Context, lv Level, msg string, args ...interface{})

	Trace(ctx context.Context, msg string, fields ...Field)
	Debug(ctx context.Context, msg string, fields ...Field)
	Info(ctx context.Context, msg string, fields ...Field)
	Warn(ctx context.Context, msg string, fields ...Field)
	Error(ctx context.Context, msg string, fields ...Field)
	Fatal(ctx context.Context, msg string, fields ...Field)
	Tracef(ctx context.Context, format string, args ...interface{})
	Debugf(ctx context.Context, format string, args ...interface{})
	Infof(ctx context.Context, format string, args ...interface{})
	Warnf(ctx context.Context, format string, args ...interface{})
	Errorf(ctx context.Context, format string, args ...interface{})
	Fatalf(ctx context.Context, format string, args ...interface{})
	Tracew(ctx context.Context, msg string, args ...interface{})
	Debugw(ctx context.Context, msg string, args ...interface{})
	Infow(ctx context.Context, msg string, args ...interface{})
	Warnw(ctx context.Context, msg string, args ...interface{})
	Errorw(ctx context.Context, msg string, args ...interface{})
	Fatalw(ctx context.Context, msg string, args ...interface{})
}

// NewLogger 创建默认的Logger
func NewLogger(config *Config) Logger {
	l := &logger{Config: config}
	if config.Async {
		l.channels = append(l.channels, NewAsyncChannel(l.channels, config.LogMax))
		l.Start()
	} else {
		l.channels = config.Channels
	}

	return l
}

type logger struct {
	*Config
	channels []Channel
}

func (l *logger) getChannel(name string) Channel {
	for _, c := range l.Channels {
		if c.Name() == name {
			return c
		}
	}

	return nil
}

func (l *logger) IsEnable(lv Level) bool {
	return lv <= l.Level
}

func (l *logger) SetLevel(name string, lv Level) {
	if name == "" {
		l.Level = lv
	} else if c := l.getChannel(name); c != nil {
		c.SetLevel(lv)
	}
}

// Start run async logger
func (l *logger) Start() {
	for _, c := range l.channels {
		c.Open()
	}
}

// Stop stop async logger
func (l *logger) Stop() {
	for _, c := range l.channels {
		c.Close()
	}
}

func (l *logger) Write(e *Entry) {
	if !l.DisableCaller {
		f := getFrame(e.CallDepth)
		e.Path = f.File
		e.File = filepath.Base(f.File)
		e.Line = f.Line
		e.Method = getFuncName(f.Function)
	}
	e.Tags = l.Tags

	for _, f := range l.Filters {
		if err := f(e); err != nil {
			return
		}
	}

	for _, c := range l.channels {
		if c.IsEnable(e.Level) {
			c.Write(e)
		}
	}
}

func (l *logger) Log(ctx context.Context, lv Level, msg string, fields ...Field) {
	if l.IsEnable(lv) {
		e := NewEntry(l)
		e.Context = ctx
		e.Level = lv
		e.Text = msg
		e.Fields = fields
		l.Write(e)
	}
}

func (l *logger) Logf(ctx context.Context, lv Level, format string, args ...interface{}) {
	if l.IsEnable(lv) {
		text := fmt.Sprintf(format, args...)
		e := NewEntry(l)
		e.Context = ctx
		e.Level = lv
		e.Text = text
		l.Write(e)
	}
}

func (l *logger) Logw(ctx context.Context, lv Level, msg string, args ...interface{}) {
	if l.IsEnable(lv) {
		e := NewEntry(l)
		e.Context = ctx
		e.Level = lv
		e.Text = msg
		if len(args) > 0 {
			e.Fields = append(e.Fields, toFields(args)...)
		}
		l.Write(e)
	}
}

func (l *logger) Trace(ctx context.Context, msg string, fields ...Field) {
	l.Log(ctx, TraceLevel, msg, fields...)
}

func (l *logger) Debug(ctx context.Context, msg string, fields ...Field) {
	l.Log(ctx, DebugLevel, msg, fields...)
}

func (l *logger) Info(ctx context.Context, msg string, fields ...Field) {
	l.Log(ctx, InfoLevel, msg, fields...)
}

func (l *logger) Warn(ctx context.Context, msg string, fields ...Field) {
	l.Log(ctx, WarnLevel, msg, fields...)
}

func (l *logger) Error(ctx context.Context, msg string, fields ...Field) {
	l.Log(ctx, ErrorLevel, msg, fields...)
}

func (l *logger) Fatal(ctx context.Context, msg string, fields ...Field) {
	l.Log(ctx, FatalLevel, msg, fields...)
}

func (l *logger) Tracef(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, TraceLevel, format, args...)
}

func (l *logger) Debugf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, DebugLevel, format, args...)
}

func (l *logger) Infof(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, InfoLevel, format, args...)
}

func (l *logger) Warnf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, WarnLevel, format, args...)
}

func (l *logger) Errorf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, ErrorLevel, format, args...)
}

func (l *logger) Fatalf(ctx context.Context, format string, args ...interface{}) {
	l.Logf(ctx, FatalLevel, format, args...)
}

func (l *logger) Tracew(ctx context.Context, msg string, args ...interface{}) {
	l.Logf(ctx, TraceLevel, msg, args...)
}

func (l *logger) Debugw(ctx context.Context, msg string, args ...interface{}) {
	l.Logf(ctx, DebugLevel, msg, args...)
}

func (l *logger) Infow(ctx context.Context, msg string, args ...interface{}) {
	l.Logf(ctx, InfoLevel, msg, args...)
}

func (l *logger) Warnw(ctx context.Context, msg string, args ...interface{}) {
	l.Logf(ctx, WarnLevel, msg, args...)
}

func (l *logger) Errorw(ctx context.Context, msg string, args ...interface{}) {
	l.Logf(ctx, ErrorLevel, msg, args...)
}

func (l *logger) Fatalw(ctx context.Context, msg string, args ...interface{}) {
	l.Logf(ctx, FatalLevel, msg, args...)
}
