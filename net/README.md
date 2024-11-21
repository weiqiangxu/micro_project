# HTTP && GRPC Server

> Gin的HTTP服务Server定义和GRPC-Server定义

1. 引入jaeger做GRPC的链路追踪
2. Gin.Server注入


### 1.trace

```bash
go.opentelemetry.io/otel
go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin
```

### 2.prometheus

```bash
github.com/prometheus/client_golang
```

### 3.pprof

```bash
github.com/gin-contrib/pprof
```

- [go-zero pprof](https://mp.weixin.qq.com/s/yYFM3YyBbOia3qah3eRVQA)

- [golang-gin实践](https://www.jishuchi.com/read/gin-practice/3833)
