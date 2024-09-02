# common-config

### logger

```
logger应该独立git仓库
热度最高的zap\logrus和官方log都是同一个抽象
参考抽象实现（我们是定义了一样的抽象但是基于zap二次封装）

logger的持久化采用异步提交elastic或者filebeat（EFK）采集或者主动push
实现了一个监听器做错误告警（企业微信、邮件）
```

### 提交代码之前必须执行

```
make fmt && make lint
```