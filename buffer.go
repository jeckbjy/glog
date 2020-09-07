package glog

import (
	"fmt"
	"strconv"
	"sync"
	"time"
)

// smallBufferSize is an initial allocation minimal capacity.
const smallBufferSize = 64
const maxInt = int(^uint(0) >> 1)

var gBufferPool = sync.Pool{
	New: func() interface{} {
		return &Buffer{buf: make([]byte, 0, smallBufferSize)}
	},
}

// NewBuffer 从缓存中获取一个Buffer
func NewBuffer() *Buffer {
	return gBufferPool.Get().(*Buffer)
}

// Buffer ...
type Buffer struct {
	buf []byte
}

// Free 将Buffer放入缓存中
func (b *Buffer) Free() {
	b.buf = b.buf[:]
	gBufferPool.Put(b)
}

// Bytes 返回Bytes数据
func (b *Buffer) Bytes() []byte {
	return b.buf
}

// String 转换为字符串
func (b *Buffer) String() string {
	return string(b.buf)
}

func (b *Buffer) At(i int) byte {
	return b.buf[i]
}

// Last 返回最后一个字符
func (b *Buffer) Last() byte {
	l := len(b.buf)
	if l > 0 {
		return b.buf[l-1]
	}

	return 0
}

// Empty 判断是否为空
func (b *Buffer) Empty() bool {
	return len(b.buf) == 0
}

// Len 返回缓存大小
func (b *Buffer) Len() int {
	return len(b.buf)
}

func (b *Buffer) Cap() int {
	return cap(b.buf)
}

func (b *Buffer) Grow(n int) {
	m := b.Len()
	if n <= cap(b.buf)-m {
		return
	}

	if b.buf == nil && n <= smallBufferSize {
		b.buf = make([]byte, n, smallBufferSize)
		return
	}

	c := cap(b.buf)
	buf := make([]byte, 2*c+n)
	copy(buf, b.buf[:])
	b.buf = buf
}

func (b *Buffer) Write(p []byte) (int, error) {
	b.buf = append(b.buf, p...)
	return len(p), nil
}

func (b *Buffer) AppendByte(v byte) {
	b.buf = append(b.buf, v)
}

func (b *Buffer) AppendString(s string) {
	b.buf = append(b.buf, s...)
}

func (b *Buffer) AppendInt(i int64) {
	b.buf = strconv.AppendInt(b.buf, i, 10)
}

func (b *Buffer) AppendUint(i uint64) {
	b.buf = strconv.AppendUint(b.buf, i, 10)
}

func (b *Buffer) AppendBool(v bool) {
	b.buf = strconv.AppendBool(b.buf, v)
}

func (b *Buffer) AppendFloat32(f float32) {
	b.buf = strconv.AppendFloat(b.buf, float64(f), 'f', -1, 32)
}

func (b *Buffer) AppendFloat64(f float64) {
	b.buf = strconv.AppendFloat(b.buf, f, 'f', -1, 64)
}

func (b *Buffer) AppendComplex64(val complex64) {
	r, i := float32(real(val)), float32(imag(val))
	b.AppendFloat32(r)
	b.AppendByte('+')
	b.AppendFloat32(i)
	b.AppendByte('i')
}

func (b *Buffer) AppendComplex128(val complex128) {
	r, i := float64(real(val)), float64(imag(val))
	b.AppendFloat64(r)
	b.AppendByte('+')
	b.AppendFloat64(i)
	b.AppendByte('i')
}

func (b *Buffer) AppendTime(t time.Time, layout string) {
	b.buf = t.AppendFormat(b.buf, layout)
}

// Appendf 带有格式化写入message
func (b *Buffer) Appendf(format string, args ...interface{}) {
	msg := fmt.Sprintf(format, args...)
	b.AppendString(msg)
}

// Putf 格式化字符串并调用Put
func (b *Buffer) Putf(min int, max int, format string, args ...interface{}) {
	text := fmt.Sprintf(format, args...)
	b.Put(min, max, text)
}

// Put 根据宽度限制写入字符串,min<0代表左对齐,>0代表右对齐，max代表输出的最大字符宽度
func (b *Buffer) Put(min int, max int, data string) {
	size := len(data)
	if size == 0 {
		return
	}

	switch {
	case max > 0 && size > max:
		b.AppendString(data[size-max:])
	case min < 0 && size < -min:
		b.Grow(-min)
		b.AppendString(data)
		b.writeSpace(-min - size)
	case min > 0 && size < min:
		b.Grow(min)
		b.writeSpace(min - size)
		b.AppendString(data)
	default:
		b.AppendString(data)
	}
}

func (b *Buffer) writeSpace(count int) {
	for i := 0; i < count; i++ {
		b.AppendByte(' ')
	}
}
