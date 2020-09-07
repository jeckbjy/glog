package glog

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
)

var (
	ErrNotReady = fmt.Errorf("channel not ready")
)

// NewFileChannel ...
func NewFileChannel(opts ...ChannelOption) Channel {
	o := NewChannelOptions(opts...)
	c := &fileChannel{}
	c.Init(o)
	return c
}

// fileChannel 输出到文件
type fileChannel struct {
	BaseChannel
	file *os.File
	path string
	err  error
}

func (c *fileChannel) Name() string {
	return "file"
}

func (c *fileChannel) Open() error {
	if c.file == nil && c.err == nil {
		if c.path == "" {
			c.path = fmt.Sprintf("%s.log", filepath.Base(os.Args[0]))
		}

		dir := path.Dir(c.path)
		err := os.MkdirAll(dir, os.ModePerm)
		if err != nil {
			c.err = err
			return err
		}
		c.file, err = os.OpenFile(c.path, os.O_RDWR|os.O_CREATE, os.ModePerm)
		if err != nil {
			c.err = err
			c.file = nil
		}
	}

	if c.file == nil {
		return ErrNotReady
	}

	return nil
}

func (c *fileChannel) Close() error {
	if c.file != nil {
		err := c.file.Close()
		c.file = nil
		return err
	}

	return nil
}

func (c *fileChannel) Write(e *Entry) {
	if c.Open() == nil {
		text := c.Format(e)
		if text != nil {
			_, _ = c.file.Write(text)
		}
	}
}
