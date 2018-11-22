压力测试工具
====

This is a stress testing util, just run `stress -h` and try `stress -url <your http url>`.

使用方法
----

```
# add $GO_PATH/bin to your $PATH
go get github.com/mapleque/stress
stress -h
stress -url 'http://localhost/hello'
```

- 支持CTRL+C暂停，Enter继续
- 设置失败率阈值，超过阈值将会自动结束
- 结束时会自动输出统计结果和错误信息

功能简介
----

本工具支持两种模式：

### 动态模式（默认）
> 动态模式下系统会持续自动增加线程，每次增加线程数由step参数决定，增加间隔时间由stay决定。    
> 其中：step默认为1个，stay默认为1s，二者必须是正整数。

### 静态模式
> 静态模式下系统会保持固定的线程数发送请求。线程数由thread决定，请求频率由interval决定。    
> 其中：
> - thread值必须是正整数，默认为1。    
> - interval必须是正整数或0，单位是毫秒，默认为1。    
> 当interval为0时，表示请求频率不限，系统会在一个线程内持续发送请求。

