package glog

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestGetFrame(t *testing.T) {
	frame := getFrame(0)
	t.Log(frame.File)
	t.Log(frame.Line)
	t.Log(frame.Function)
}

func TestPrint(t *testing.T) {
	ss := fmt.Sprintf("%-10s aa", "aa")
	t.Log(ss)
}

func TestExecName(t *testing.T) {
	t.Log(filepath.Base(os.Args[0]))
}

func TestDateFormat(t *testing.T) {
	df := DateFormat{}
	df.Parse("yyyy-MM-ddTHH:mm:ssz")
	str := df.Format(time.Now())
	t.Log(str)
}

func TestLayout(t *testing.T) {
	l := &Layout{}
	if err := l.Parse("[%d{yyyy-MM-dd HH:mm:ss.fff}][%p] (%F:%L) - %m%f%n"); err != nil {
		t.Error(err)
	}
	msg := &Entry{
		Time:  time.Now(),
		Level: DebugLevel,
		Text:  "test layout",
		File:  "test",
	}

	s := string(l.Format(msg))
	t.Log(s)
}

func TestLogging(t *testing.T) {
	Tracef(nil, "%s %s %s", "hello", "world", "trace")
	Debugf(nil, "%s %s %s", "hello", "world", "debug")
	Infof(nil, "%s %s %s", "hello", "world", "info")
	Warnf(nil, "%s %s %s", "hello", "world", "warn")
	Errorf(nil, "%s %s %s", "hello", "world", "error")
	Fatalf(nil, "%s %s %s", "hello", "world", "fatal")

	Trace(nil, "hello world", Int("status", 404), String("method", "Get"))
}

func TestGraylog(t *testing.T) {
	conf := NewConfig()
	conf.AddTags(map[string]string{
		"facility": "staging_test",
	})

	url := "testing url"
	channel := NewGraylogChannel(WithURL(url))
	conf.AddChannels(channel)

	logger := NewLogger(conf)
	logger.Infof(nil, "test glog")
}
