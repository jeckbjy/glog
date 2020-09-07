package glog

import (
	"bytes"
	"fmt"
	"net/http"
	"path"
	"time"
)

// NewElasticChannel ...
func NewElasticChannel(opts ...ChannelOption) Channel {
	o := NewChannelOptions(opts...)
	c := &elasticChannel{}
	c.Init(o)
	c.client = o.HttpClient
	c.Retry = o.Retry
	c.Bulk = o.Batch
	c.Index = o.IndexName
	return c
}

// 官方的API: https://github.com/elastic/go-elasticsearch
// 但是比较厚重,这里只需要Index
// https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-index_.html
// https://www.elastic.co/guide/en/elasticsearch/reference/current/docs-bulk.html
// TODO:支持Bulk模式,需要ndjson编码
// formatter编码格式必须是json格式
type elasticChannel struct {
	BaseChannel
	URL    string       // 地址
	Retry  int          // 失败重试次数,默认不重试,直接丢弃
	Bulk   int          // 用于配置一个批次发送多少条日志,默认1条
	Index  string       // 索引名,按照日期分类?
	client *http.Client //
}

func (c *elasticChannel) Name() string {
	return "elastic"
}

func (c *elasticChannel) Open() error {
	if c.client == nil {
		c.client = &http.Client{Timeout: time.Second * 10}
	}

	return nil
}

func (c *elasticChannel) Close() error {
	if c.client != nil {
		c.client.CloseIdleConnections()
		c.client = nil
	}

	return nil
}

// 处理速度可能比较慢,放到单独一个队列中处理
func (c *elasticChannel) Write(msg *Entry) {
	if c.Open() != nil {
		return
	}

	// POST /<index>/_doc/
	// POST /<index>/_create/<_id>
	text := c.Format(msg)
	url := path.Join(c.URL, c.getIndex(), "_doc")
	_ = c.doPost(url, "application/json", text)
}

// WriteBatch 批量发送
func (c *elasticChannel) WriteBatch(q *Queue) {

}

// 根据当前时间按天进行索引
func (c *elasticChannel) getIndex() string {
	now := time.Now()
	index := fmt.Sprintf("%s_%04d%02d%02d", c.Index, now.Year(), now.Month(), now.Day())
	return index
}

// POST /_bulk
// POST /<index>/_bulk
// TODO:support ndjson
// func (c *elasticChannel) doSendBulk(msg []*Entry) {
// 	//url := path.Join(c.URL, c.getIndex(), "_bulk")
// 	//c.doPost(url, "application/x-ndjson")
// }

// 添加重试功能,失败则丢弃
func (c *elasticChannel) doPost(url, contentType string, data []byte) error {
	var err error
	for i := 0; i < c.Retry+1; i++ {
		_, err = c.client.Post(url, contentType, bytes.NewReader(data))
		if err == nil {
			return nil
		}
	}

	return err
}
