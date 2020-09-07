# glog golang日志库

有很多优秀的开源日志库，比如logrus,zap,他们都是基于structured log思想实现,其中zap效率更高,logrus开发更早，使用更广泛。之所以重复造轮子，主要是可以更灵活的控制

## 与logrus,zap的一些差异
- 提供了一个异步队列,在开发环境可以使用同步输出，在线上可以使用异步输出,但可能会丢失日志
- 所有输出都是对等一个Channel,而不是logrus中的Writer+Hook的模式,提供了几个常见的channel，包括console,file,graylog,AsyncChannel
- 提供了一个类似Log4j的Layout输出格式解析
- 增加Tags信息,用于log初始化时设置env,host,idc,facility,psm,cluster,pod,stage,unit等信息
- 对于context.Context处理,在很多RPC服务中,通常会透传context,在打印日志时,第一个参数通常会传入ctx,调用者通常会通过Context向Fileds中写入RequestID等信息，对日志系统而言本身并不知道如何处理context,可以配合Filter设置相关Field

## 使用方法
```go
func main() {
    // 初始化channel和相关配置
    conf := NewConfig()
	conf.AddTags(map[string]string{
		"facility": "staging_test",
	})

	url := "testing url"
	channel := NewGraylogChannel(WithURL(url))
	conf.AddChannels(channel)

    logger := NewLogger(conf)
    // 替换默认logger
    SetDefault(logger)
	logger.Infof(nil, "test glog")
}
```

## TODO
- 测试graylog,elastic
- 完善file rotate

## 
- [zap](https://github.com/uber-go/zap)
- [logrus](https://github.com/sirupsen/logrus)
- [klog](https://github.com/kubernetes/klog)
- [Go logs](https://www.ctolib.com/topics-123640.html)
- [Log4j PatternLayout](https://wiki.jikexueyuan.com/project/log4j/log4j-patternlayout.html)
- [获取函数名](https://colobu.com/2018/11/03/get-function-name-in-go/)
- [runtime.Caller性能](https://cloud.tencent.com/developer/article/1385947)
- [color](https://github.com/gookit/color)