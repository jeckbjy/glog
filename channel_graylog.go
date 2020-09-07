package glog

import (
	"bytes"
	"compress/flate"
	"compress/gzip"
	"compress/zlib"
	"crypto/rand"
	"fmt"
	"io"
	"net"
	"strings"
)

const (
	chunkedSize    = 1420 // TODO:通过MTU确定?
	chunkedHeadLen = 12
	chunkedDataLen = chunkedSize - chunkedHeadLen
)

var (
	magicChunked = []byte{0x1e, 0x1f}
	magicZlib    = []byte{0x78}
	magicGzip    = []byte{0x1f, 0x8b}
)

var (
	graylogTcpDelimited = []byte{0}
)

// NewGraylogChannel ...
func NewGraylogChannel(opts ...ChannelOption) Channel {
	o := NewChannelOptions(opts...)
	c := &graylogChannel{}
	c.Init(o)
	c.compressLevel = o.CompressLevel
	c.compressType = o.CompressType
	if c.compressLevel == -1 {
		c.compressLevel = flate.BestSpeed
	}
	c.localIP = o.LocalIP
	if c.localIP == "" {
		c.localIP = getLocalIp()
	}

	c.network = "udp"
	c.addr = o.URL
	if index := strings.Index(o.URL, "://"); index != -1 {
		c.network = o.URL[:index]
		c.addr = o.URL[index+3:]
	}

	c.Open()
	return c
}

// https://docs.graylog.org/en/2.4/pages/gelf.html
// https://docs.graylog.org/en/2.1/pages/gelf.html
// https://en.wikipedia.org/wiki/Syslog
// GELF支持UDP,TCP,HTTP
type graylogChannel struct {
	BaseChannel
	localIP       string       // host
	network       string       // tcp/udp
	addr          string       // Graylog连接地址
	compressLevel int          // 压缩级别
	compressType  CompressType // 压缩类型
	conn          net.Conn
}

func (c *graylogChannel) Name() string {
	return "graylog"
}

func (c *graylogChannel) Open() error {
	if c.conn == nil && c.addr != "" {
		if conn, err := net.Dial(c.network, c.addr); err != nil {
			fmt.Printf("open graylog fail,%+v", c.addr)
			return err
		} else {
			c.conn = conn
		}
	}

	return nil
}

func (c *graylogChannel) Close() error {
	if c.conn != nil {
		c.conn.Close()
		c.conn = nil
	}

	return nil
}

func (c *graylogChannel) Write(e *Entry) {
	if c.conn == nil {
		return
	}

	enc := NewJsonEncoder()
	enc.Begin()
	enc.AddString("version", "1.1")
	enc.AddString("host", c.localIP)
	enc.AddString("short_message", e.Text)
	enc.AddString("full_message", e.Text)
	enc.AddFloat64("timestamp", float64(e.Time.UnixNano()/1000000)/1000.)
	enc.AddInt("level", int64(e.Level.ToSyslogLevel()))

	if e.File != "" {
		enc.AddString("_file", e.File)
		enc.AddInt("_line", int64(e.Line))
	}
	enc.AddValidString("_method", e.Method)

	for i := 0; i < e.Tags.Len(); i++ {
		k, v := e.Tags.GetAt(i)
		enc.AddString("_"+k, v)
	}

	for _, f := range e.Fields {
		enc.AddField(&f, toGraylogExtraKey)
	}
	enc.End()
	switch c.network {
	case "tcp":
		// GELF TCP不支持压缩,以\0分隔
		// 同步阻塞发送
		data := enc.Bytes()
		c.conn.Write(data)
		c.conn.Write(graylogTcpDelimited)
	case "udp":
		data := c.compress(enc.Bytes())
		var err error
		if len(data) <= chunkedSize {
			_, err = c.conn.Write(data)
		} else {
			err = c.writeChunked(data)
		}
		if err != nil {
			fmt.Printf("send graylog fail,%+v", err)
		}
	}
}

func (c *graylogChannel) writeChunked(data []byte) error {
	n := len(data)/chunkedDataLen + 1
	if n > 128 {
		return fmt.Errorf("msg too large, would need %d chunks", n)
	}
	chunkNum := uint8(n)
	// use random to get a unique message id
	msgID := make([]byte, 8)
	n, err := io.ReadFull(rand.Reader, msgID)
	if err != nil || n != 8 {
		return fmt.Errorf("rand.Reader: %d/%s", n, err)
	}

	b := make([]byte, 0, chunkedSize)
	buf := bytes.NewBuffer(b)

	bytesLeft := len(data)
	for i := uint8(0); i < chunkNum; i++ {
		buf.Reset()
		// manually write header.  Don't care about
		// host/network byte order, because the spec only
		// deals in individual bytes.
		buf.Write(magicChunked) //magic
		buf.Write(msgID)
		buf.WriteByte(i)
		buf.WriteByte(chunkNum)
		// slice out our chunk from zBytes
		chunkLen := chunkedDataLen
		if chunkLen > bytesLeft {
			chunkLen = bytesLeft
		}
		off := int(i) * chunkedDataLen
		chunk := data[off : off+chunkLen]
		buf.Write(chunk)

		// write this chunk, and make sure the write was good
		n, err := c.conn.Write(buf.Bytes())
		if err != nil {
			return fmt.Errorf("Write (chunk %d/%d): %s", i, chunkNum, err)
		}
		if n != len(buf.Bytes()) {
			return fmt.Errorf("Write len: (chunk %d/%d) (%d/%d)", i, chunkNum, n, len(buf.Bytes()))
		}

		bytesLeft -= chunkLen
	}

	if bytesLeft != 0 {
		return fmt.Errorf("error: %d bytes left after sending", bytesLeft)
	}
	return nil
}

func (c *graylogChannel) compress(data []byte) []byte {
	fmt.Printf("%s\n", data)
	var buf *Buffer
	var w io.WriteCloser
	var err error
	switch c.compressType {
	case CompressGzip:
		buf = NewBuffer()
		w, err = gzip.NewWriterLevel(buf, c.compressLevel)
	case CompressZlib:
		buf = NewBuffer()
		w, err = zlib.NewWriterLevel(buf, c.compressLevel)
	default:
		return data
	}

	buf.Grow(len(data))
	if _, err = w.Write(data); err != nil {
		w.Close()
		return nil
	}
	w.Close()
	return buf.Bytes()
}

func toGraylogExtraKey(key string) string {
	return "_" + key
}
