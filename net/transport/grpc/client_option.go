package grpc

import (
	"github.com/opentracing/opentracing-go"
	"google.golang.org/grpc"
)

// clientOptions is gRPC Client
type clientOptions struct {
	endpoint           string
	unaryInterceptors  []grpc.UnaryClientInterceptor
	streamInterceptors []grpc.StreamClientInterceptor
	grpcOpts           []grpc.DialOption
	tracing            bool
	insecure           bool
	prometheus         bool
	tracer             opentracing.Tracer
	tracerInterceptor  bool
}

// ClientOption is gRPC client option.
type ClientOption func(o *clientOptions)

// WithEndpoint with client endpoint.
func WithEndpoint(endpoint string) ClientOption {
	return func(c *clientOptions) {
		c.endpoint = endpoint
	}
}

// WithUnaryTraceInterceptor with client endpoint.
func WithUnaryTraceInterceptor(tracer opentracing.Tracer) ClientOption {
	return func(c *clientOptions) {
		// 注入Jaeger追踪器Tracer
		c.tracer = tracer
		c.tracerInterceptor = true
	}
}

// WithPrometheus prometheus metrics
func WithPrometheus(b bool) ClientOption {
	return func(c *clientOptions) {
		c.prometheus = b
	}
}

// WithUnaryInterceptor returns a DialOption that specifies the interceptor for unary RPCs.
func WithUnaryInterceptor(u ...grpc.UnaryClientInterceptor) ClientOption {
	return func(c *clientOptions) {
		c.unaryInterceptors = append(c.unaryInterceptors, u...)
	}
}

// WithOptions with gRPC options.
func WithOptions(opts ...grpc.DialOption) ClientOption {
	return func(c *clientOptions) {
		c.grpcOpts = append(c.grpcOpts, opts...)
	}
}

// WithTracing trace
func WithTracing(opt bool) ClientOption {
	return func(c *clientOptions) {
		c.tracing = opt
	}
}

// WithInSecure insecure is true will create client no secure
func WithInSecure(insecure bool) ClientOption {
	return func(c *clientOptions) {
		c.insecure = insecure
	}
}
