package grpc

import (
	"context"
	"fmt"
	"github.com/weiqiangxu/micro_project/user/global/enum"
	"google.golang.org/grpc/keepalive"
	"time"

	grpcPrometheus "github.com/grpc-ecosystem/go-grpc-prometheus"
	"github.com/opentracing/opentracing-go"
	prom "github.com/prometheus/client_golang/prometheus"
	"github.com/weiqiangxu/micro_project/user/config"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/balancer/roundrobin"
	grpcInsecure "google.golang.org/grpc/credentials/insecure"
)

// Dial rpc client dial an address
func Dial(ctx context.Context, opts ...ClientOption) (*grpc.ClientConn, error) {
	options := clientOptions{}
	for _, o := range opts {
		o(&options)
	}
	if options.tracing {
		// otelgrpc.UnaryClientInterceptor() 启动一个 OpenTelemetry 追踪跨度（Span）
		// 追踪跨度代表了一个工作单元，在这里就是 gRPC 客户端请求的整个生命周期，从发送请求开始，到收到响应结束
		// 拦截器会将 OpenTelemetry 的追踪上下文信息（如追踪 ID、跨度 ID 等）注入到 gRPC 请求的元数据（Metadata）中
		// 服务端在收到请求后，可以根据这些元数据来继续这个追踪跨度或者创建新的相关跨度，从而实现端到端的请求链路追踪
		options.unaryInterceptors = append(options.unaryInterceptors, otelgrpc.UnaryClientInterceptor())
		options.streamInterceptors = append(options.streamInterceptors, otelgrpc.StreamClientInterceptor())
	}
	grpcOpts := []grpc.DialOption{
		grpc.WithDefaultServiceConfig(fmt.Sprintf(`{"LoadBalancingPolicy": %q}`, roundrobin.Name)),
		grpc.WithChainUnaryInterceptor(options.unaryInterceptors...),
		grpc.WithChainStreamInterceptor(options.streamInterceptors...),
	}
	// 开启Prometheus指标记录RPC调用的时长
	if options.prometheus {
		grpcPrometheus.EnableClientHandlingTimeHistogram(
			WithGrpcHistogramName(config.Conf.Application.Name, "grpc_seconds"),
		)
		list := []grpc.DialOption{
			grpc.WithChainUnaryInterceptor(grpcPrometheus.UnaryClientInterceptor),
			grpc.WithChainStreamInterceptor(grpcPrometheus.StreamClientInterceptor),
		}
		grpcOpts = append(grpcOpts, list...)
	}
	if options.insecure {
		// RPC如果开启了TLS/SSL
		grpcOpts = append(grpcOpts, grpc.WithTransportCredentials(grpcInsecure.NewCredentials()))
	}
	// RPC调用时候记录链路追踪指标
	if options.tracerInterceptor {
		list := []grpc.DialOption{
			// 将Jaeger 追踪器（Tracer）注入一元RPC调用拦截器
			grpc.WithChainUnaryInterceptor(ClientInterceptor(options.tracer)),
			grpc.WithChainUnaryInterceptor(ReleaseInterceptor()),
			grpc.WithChainUnaryInterceptor(ReleaseInterceptorNew()),
		}
		grpcOpts = append(grpcOpts, list...)
	}
	if len(options.grpcOpts) > 0 {
		grpcOpts = append(grpcOpts, options.grpcOpts...)
	}
	keepAliveOpt := grpc.WithKeepaliveParams(keepalive.ClientParameters{
		// 即每隔 10 秒发送一次 Keep-Alive 消息
		Time: 10 * time.Second,
		// 如果在3秒内没有收到响应就认为连接可能出现问题
		Timeout:             3 * time.Second,
		PermitWithoutStream: true,
	})
	grpcOpts = append(grpcOpts, keepAliveOpt)

	return grpc.DialContext(ctx, options.endpoint, grpcOpts...)
}

// WithGrpcHistogramName change prometheus histogramName
func WithGrpcHistogramName(namespace string, name string) grpcPrometheus.HistogramOption {
	return func(o *prom.HistogramOpts) {
		o.Namespace = namespace
		o.Name = name
	}
}

// ClientInterceptor 记录RPC调用这个Span的执行时长
func ClientInterceptor(tracer opentracing.Tracer) grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 实现一个一元调用拦截器
		// 从上下文之中获取父跨度标识
		fmt.Println("ClientInterceptor")
		parentSpan := ctx.Value(enum.TraceSpanName)
		if parentSpan != nil {
			parentSpanContext := parentSpan.(opentracing.SpanContext)
			// 在当前的http请求跨度下面创建一个子span也就是子跨度
			// 使用Jaeger 追踪器（Tracer）创建子跨度span
			child := tracer.StartSpan(
				fmt.Sprintf("grpc.request:%s", method),
				opentracing.ChildOf(parentSpanContext))
			// 调用完成以后标识此Span结束
			defer child.Finish()
		}
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// ReleaseInterceptor RPC调用完成释放连接
func ReleaseInterceptor() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 实现一个一元调用拦截器
		fmt.Println("GRPC请求开始...")
		defer func() {
			fmt.Println("GRPC的请求完成通知gRPC连接池释放连接...")
		}()
		// invoker真正的发起RPC调用,并且调用完成之后进入defer
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}

// ReleaseInterceptorNew RPC调用完成释放连接
func ReleaseInterceptorNew() grpc.UnaryClientInterceptor {
	return func(ctx context.Context, method string, req, reply interface{}, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
		// 实现一个一元调用拦截器
		fmt.Println("GRPC请求开始New...")
		defer func() {
			fmt.Println("GRPC的请求完成通知gRPC连接池释放连接New...")
		}()
		// invoker真正的发起RPC调用,并且调用完成之后进入defer
		return invoker(ctx, method, req, reply, cc, opts...)
	}
}
