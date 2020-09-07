package glog

import "net/http"

const (
	// CompressNone .
	CompressNone CompressType = iota
	// CompressGzip .
	CompressGzip
	// CompressZlib .
	CompressZlib
)

type CompressType int

// ChannelOptions Channel常见可选配置
type ChannelOptions struct {
	Level         Level        // 日志级别
	Layout        string       // Text Format输出格式
	Formatter     Formatter    // 输出格式
	File          string       // 文件输出路径
	URL           string       // 连接用URL,支持scheme为tcp或udp
	LocalIP       string       // 本地地址
	CompressLevel int          // 压缩级别
	CompressType  CompressType // 压缩类型
	HttpClient    *http.Client // elastic使用http协议
	Retry         int          // 重试次数
	Batch         int          // 一次发送大小
	IndexName     string       // elastic索引名
}

type ChannelOption func(o *ChannelOptions)

// NewChannelOptions ...
func NewChannelOptions(opts ...ChannelOption) *ChannelOptions {
	o := &ChannelOptions{Level: TraceLevel}
	for _, fn := range opts {
		fn(o)
	}

	if o.Formatter == nil {
		if o.Layout != "" {
			o.Formatter, _ = NewTextFormatter(o.Layout)
		}
		if o.Formatter == nil {
			o.Formatter = DefaultFormatter
		}
	}

	o.CompressLevel = -1
	o.CompressType = CompressNone

	return o
}

func WithLevel(lv Level) ChannelOption {
	return func(o *ChannelOptions) {
		o.Level = lv
	}
}

func WithLayout(l string) ChannelOption {
	return func(o *ChannelOptions) {
		o.Layout = l
	}
}

func WithFormatter(f Formatter) ChannelOption {
	return func(o *ChannelOptions) {
		o.Formatter = f
	}
}

func WithURL(url string) ChannelOption {
	return func(o *ChannelOptions) {
		o.URL = url
	}
}

func WithLocalIP(ip string) ChannelOption {
	return func(o *ChannelOptions) {
		o.LocalIP = ip
	}
}

func WithCompressType(ct CompressType) ChannelOption {
	return func(o *ChannelOptions) {
		o.CompressType = ct
	}
}

func WithCompressLevel(level int) ChannelOption {
	return func(o *ChannelOptions) {
		o.CompressLevel = level
	}
}

func WithCompress(ct CompressType, level int) ChannelOption {
	return func(o *ChannelOptions) {
		o.CompressType = ct
		o.CompressLevel = level
	}
}

func WithRetry(retry int) ChannelOption {
	return func(o *ChannelOptions) {
		o.Retry = retry
	}
}

func WithHttpClient(c *http.Client) ChannelOption {
	return func(o *ChannelOptions) {
		o.HttpClient = c
	}
}

func WithBatch(b int) ChannelOption {
	return func(o *ChannelOptions) {
		o.Batch = b
	}
}

func WithIndexName(index string) ChannelOption {
	return func(o *ChannelOptions) {
		o.IndexName = index
	}
}
