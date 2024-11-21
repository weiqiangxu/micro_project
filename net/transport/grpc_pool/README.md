# grpc pool

> GRPC连接池

### 一、设计逻辑

##### 1.数量控制

1. MinActiveConnections 最小连接
	服务启动的时候就会创建这个数据量的连接，并且定时保活，即使没有发起任何的grpc请求到服务端。
2. MaxActiveConnections 最大连接
	连接池中允许同时存在的正在被用于处理请求的最大连接数量，grpc请求频繁的时候连接会增加，但最大数量不会超过最大活跃连接。
3. MaxIdleConnections 最大空闲连接
	连接池中允许存在的处于空闲状态的最大连接数量，防止太多空闲连接浪费内存，在grpc请求完成后会释放连接到连接池，此时会检查连接池的连接数量，如果超过这个数量那么会释放连接。


### 二、设计逻辑

##### 1.保活

- GRPC客户端通过设置keepAlive参数开启保活功能

```go
// google.golang.org/grpc
// google.golang.org/grpc/keepalive
opts := []grpc.DialOption{
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		// 即每隔 10 秒发送一次 Keep-Alive 消息
		Time:                10 * time.Second,
		// 如果在3秒内没有收到响应就认为连接可能出现问题
		// 当Keepalive的响应超时后Go-GRPC库会自动关闭出现问题的连接
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	}),
}
conn, err := grpc.Dial(address, opts...)
```

> 由于GRPC自带的保活的keep-alive在超时响应后的事件没办法将事件传递给连接池，所以连接池需要自定义定时器检查连接状态

- 服务端保活连接

```go
import (
    "google.golang.org/grpc"
    "google.golang.org/grpc/keepalive"
)
s := grpc.NewServer(
    grpc.KeepaliveParams(keepalive.ServerParameters{
		// 连接最大空闲时间，超过这个时间如果没有数据传输，连接可能会被关闭（默认值是无限（infinity））
		// 默认不会因为空闲时间过长而自动关闭连接
		MaxConnectionIdle: 15 * time.Hour,
		// 连接的最大存活时间，超过这个时间连接也会被关闭(默认值是无限（infinity）)
		// 默认连接不会因为存活时间过长而自动关闭
		MaxConnectionAge: 30 * time.Minute,
		// 服务端每隔10秒钟发送 Keep-Alive 消息(默认值是 2 小时)
		Time: 10 * time.Second,
		// 超过3秒钟没收到 keep-alive 的消息认为此连接无效(默认值是 20 秒)
		Timeout: 3 * time.Second,
	})
    grpc.KeepaliveEnforcementPolicy(keepalive.EnforcementPolicy{
        MinTime:             5 * time.Second,
        MaxTime:             15 * time.Second,
        PermitWithoutStream: true,
    }),
)
```

```go
// 连接池定时检查连接状态
// 创建一个定时器，设置时间间隔为10秒
ticker := time.NewTicker(10 * time.Second)

// 使用一个无限循环来持续监听定时器的触发
for range ticker.C {
	ClientConn.ClientConnGetState() connectivity.State
}

// 如果连接是"SHUTDOWN"连接已经被关闭或者正在被关闭那么就释放连接.
```

##### 2.扩容

发起RPC调用的时候运行`grpcPool.Get()`时候，从连接池之中获取`空闲`的连接，如果没有的话，那么可能增加连接。怎么判断连接是`空闲`的呢？

##### 3.缩容

> 什么时候连接池的连接会被释放

```go
// 调用连接释放函数
func (cc *ClientConn) Close() error
```

```go
// 每次Get连接的时候检查所有空闲状态下的连接数量
// 如果超过了最大空闲连接数就释放连接
ClientConn.ClientConnGetState()

// google.golang.org\grpc@v1.68.0\clientconn.go
// GetState returns the connectivity.State of ClientConn.
func (cc *ClientConn) GetState() connectivity.State {
	return cc.csMgr.getState()
}

type State int

func (s State) String() string {
	switch s {
	case Idle:
		return "IDLE" // 空闲
	case Connecting:
		return "CONNECTING" // 连接正在建立过程中
	case Ready:
		return "READY" // 连接已经成功建立并且可以用于发送和接收 RPC 请求和响应
	case TransientFailure:
		return "TRANSIENT_FAILURE" // 连接出现了暂时的故障 gRPC 库尝试恢复操作，比如重新连接、重试请求等
	case Shutdown:
		return "SHUTDOWN" // 连接已经被关闭或者正在被关闭
	default:
		logger.Errorf("unknown connectivity state: %d", s)
		return "INVALID_STATE"
	}
}
```

> 通过互斥锁保证并发安全，使用channel存储连接对象保证。

### 三、GRPC底层优势

1. grpc-go的底层http2支持多路复用现在因为连接池变成了多请求直接不是直接复用连接
2. HTTP/2的多路复用指的是多个请求公用一个TCP连接可以乱序，一个http请求表示一个流，数据包分成帧，每个帧有流id，多个流的帧可以同时在1 个 TCP 连接上传输，所以帧（也就是数据包）的到达顺序可能是乱序的，不会因为1个http请求阻塞影响其他的http请求。
3. Protocol Buffers（简称 Protobuf）序列化数据结构的方法,非常紧凑的二进制格式来存储和传输数据

### 参考

##### 1.Java的线程池

> 参考Java的线程池做GRPC的连接池

- [JavaGuide线程池](https://javaguide.cn/java/concurrent/java-thread-pool-summary.html)

1. corePoolSize 线程池的核心线程数量
2. maximumPoolSize 线程池的最大线程数
3. keepAliveTime 当线程数大于核心线程数时多余的空闲线程存活的最长时间

> 注意:其线程池的释放是在用的超时机制

##### 2.开源的GRPC连接池

- [Github-Golang实现的连接池](https://github.com/silenceper/pool/blob/master/README_ZH_CN.md)