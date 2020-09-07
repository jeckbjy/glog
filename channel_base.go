package glog

// BaseChannel 默认实现
type BaseChannel struct {
	level     Level
	formatter Formatter
}

func (c *BaseChannel) Init(o *ChannelOptions) {
	c.level = o.Level
	c.formatter = o.Formatter
}

func (c *BaseChannel) IsEnable(lv Level) bool {
	return c.level <= lv
}

func (c *BaseChannel) Level() Level {
	return c.level
}

func (c *BaseChannel) SetLevel(lv Level) {
	c.level = lv
}

func (c *BaseChannel) Open() error {
	return nil
}

func (c *BaseChannel) Close() error {
	return nil
}

func (c *BaseChannel) Format(e *Entry) []byte {
	e.Lock()
	defer e.Unlock()
	if data, ok := e.outputs[c.formatter]; ok {
		return data
	}
	data, err := c.formatter.Format(e)
	if err != nil {
		return nil
	}
	e.outputs[c.formatter] = data
	return data
}
