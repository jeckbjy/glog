package glog

import "sync"

const (
	statusNone    = 0 // 尚未运行
	statusRunning = 1 // 运行中
	statusStop    = 2 // 已经结束
)

// NewAsyncChannel 创建异步队列
func NewAsyncChannel(channels []Channel, logMax int) Channel {
	c := &asyncChannel{
		status: statusNone,
		logMax: logMax,
	}
	c.cond = sync.NewCond(&c.mux)
	for _, channel := range channels {
		if bc, ok := channel.(BatchChannel); ok {
			c.batches = append(c.batches, bc)
		} else {
			c.channels = append(c.channels, channel)
		}
	}

	return c
}

// asyncChannel 异步队列,超过LogMax的消息会被丢弃
// 如果Channel实现了Batch接口,则会调用批量发送接口
type asyncChannel struct {
	BaseChannel
	channels []Channel
	batches  []BatchChannel
	mux      sync.Mutex
	cond     *sync.Cond
	queue    Queue
	status   int
	logMax   int
}

func (c *asyncChannel) Name() string {
	return "async"
}

func (c *asyncChannel) Open() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.status == statusNone {
		c.status = statusRunning
		go c.Run()
	}
	return nil
}

func (c *asyncChannel) Close() error {
	c.mux.Lock()
	defer c.mux.Unlock()
	if c.status == statusRunning {
		c.status = statusStop
		c.cond.Signal()
	}

	for _, ch := range c.channels {
		ch.Close()
	}

	return nil
}

func (c *asyncChannel) Write(e *Entry) {
	c.mux.Lock()
	notify := false
	if c.status == statusRunning && c.queue.Len() < c.logMax {
		c.queue.Push(e)
		notify = true
	}
	c.mux.Unlock()
	if notify {
		c.cond.Signal()
	}
}

func (l *asyncChannel) Run() {
	for {
		l.mux.Lock()
		for l.status != statusStop && l.queue.Empty() {
			l.cond.Wait()
		}
		quit := l.status == statusStop
		queue := l.queue
		l.queue.Clear()
		l.mux.Unlock()

		if len(l.batches) > 0 {
			mapper := newBatchMapper()
			for _, c := range l.batches {
				batch := mapper.GetBatch(&queue, c)
				if len(batch) > 0 {
					c.WriteBatch(batch)
				}
			}
		}

		for {
			e := queue.Pop()
			if e == nil {
				break
			}

			for _, c := range l.channels {
				if c.IsEnable(e.Level) {
					c.Write(e)
				}
			}

			// 使用完释放掉
			e.Free()
		}

		queue.Clear()
		if quit {
			break
		}
	}
}

type batchMapper struct {
	batchMap map[Level][]*Entry
}

func newBatchMapper() batchMapper {
	return batchMapper{batchMap: make(map[Level][]*Entry)}
}

func (b *batchMapper) GetBatch(q *Queue, c BatchChannel) []*Entry {
	if res, ok := b.batchMap[c.Level()]; ok {
		return res
	}

	res := make([]*Entry, 0, q.Len())
	iter := q.Iterator()
	for iter.HasNext() {
		e := iter.Next()
		if c.IsEnable(e.Level) {
			res = append(res, e)
		}
	}

	b.batchMap[c.Level()] = res
	return res
}
