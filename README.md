# micro_project

> 使用Gin/GRPC搭建服务

### 一、功能列表

1. `protocol`定义grpc接口并且实现`grpc-server`和`grpc-client`
2. `prometheus`指标输出http请求时延
3. `jaeger`链路追踪在grpc之中
4. `grpc`连接池
5. 自定义logger统一日志输出方式和格式
6. 统一MQ接口并且实现使用`confluent-kafka`作为实现
7. 引入`golang-jwt`进行身份验证和授权
8. 使用`gorm.io/gorm`启动`MySQL`客户端
9. 使用`go.mongodb.org`创建`MongoDB`客户端
10. 开启`pprof`支持性能分析
11. 领域驱动设计分层


### 二、启动服务

```bash
# 启动grpc服务
$ go run user/cmd/server

# 启动grpc客户端
$ go run user/cmd/client

# 访问Http服务验证client从grpc-server读取数据
$ curl http://127.0.0.1:8989/user/info
```


### 附录

##### 1.rpc调试工具

- [bloomrpc](https://github.com/bloomrpc/bloomrpc)
- [grpcui](https://github.com/fullstorydev/grpcui)