package application

import (
	"context"
	"fmt"
	"github.com/uber/jaeger-client-go"
	"github.com/weiqiangxu/micro_project/common-config/format"
	"io"
	"reflect"
	"time"

	"github.com/opentracing/opentracing-go"
	jaegerConfig "github.com/uber/jaeger-client-go/config"

	redisApi "github.com/weiqiangxu/micro_project/common-config/cache"
	"github.com/weiqiangxu/micro_project/common-config/logger"
	"github.com/weiqiangxu/micro_project/net/transport"
	"github.com/weiqiangxu/micro_project/net/transport/grpc"
	pbUser "github.com/weiqiangxu/micro_project/protocol/user"
	adminGrpc "github.com/weiqiangxu/micro_project/user/application/admin_service/grpc"
	"github.com/weiqiangxu/micro_project/user/application/event"
	frontHttp "github.com/weiqiangxu/micro_project/user/application/front_service/http"
	"github.com/weiqiangxu/micro_project/user/config"
	"github.com/weiqiangxu/micro_project/user/domain/user"
)

var App app

type app struct {
	FrontService *frontService
	AdminService *adminService
	Event        []transport.Server
	Tracer       opentracing.Tracer
}

type frontService struct {
	UserHttp *frontHttp.UserAppHttpService
}

type adminService struct {
	UserGrpcService *adminGrpc.UserAppGrpcService
}

func Init() {

	// 创建一个 Jaeger 追踪器（Tracer）
	tracer, _ := InitJaeger(fmt.Sprintf("%s:%s",
		config.Conf.Application.Name,
		config.Conf.Application.Version))

	var loginClient pbUser.LoginClient
	if !reflect.DeepEqual(config.Conf.UserGrpcConfig, format.GrpcConfig{}) {
		// 如果是客户端才需要连接
		userGrpcConn, err := grpc.Dial(
			context.Background(),
			grpc.WithInSecure(true),
			grpc.WithEndpoint(config.Conf.UserGrpcConfig.Addr),
			grpc.WithTracing(true),
			grpc.WithPrometheus(true),
			grpc.WithUnaryTraceInterceptor(tracer),
		)
		if err != nil {
			logger.Fatal(err)
		}
		// 创建一个定时器，设置时间间隔为5秒（可根据需求修改）
		ticker := time.NewTicker(60 * time.Second)
		// 使用for循环来持续接收定时器的触发事件
		go func() {
			for range ticker.C {
				logger.Info("当前GRPC连接状态:", userGrpcConn.GetState().String())
			}
		}()

		// TODO 获取userGrpcConn的方式更改为GRPC连接池获取
		// userGrpcConn := grpcPool.Get()
		// TODO grpc连接池获取连接的时候怎么判断这个连接目前是空闲的呢,请求发起到请求完成有事件吗
		// (拦截器的defer是可以接收到完成)
		// TODO grpc连接池获取连接的时候这个连接的顺序是怎么保证的呢

		// TODO 连接池代码和接口调用代码的耦合比较严重,因为连接没有状态标识是否空闲
		// TODO 可以把 Dial时候传入pool的key 在请求完成的事件将 key通知到 pool从而释放连接
		// 请求完成的事件放在拦截器的 defer 事件之中
		// currentConn := pool.GetGrpcConn() // 计数器+1
		// pbUser.NewLoginClient(currentConn.GetConn()).ListUser() // RPC调用获取数据
		// defer currentConn.Release() // 释放回GRPC连接池

		loginClient = pbUser.NewLoginClient(userGrpcConn)
	}

	// inject rpc client && redis into domain service
	redis := redisApi.NewRedisApi(config.Conf.WikiRedisDb)
	userDomain := user.NewUserService(user.WithRedis(redis))
	frontSrv := &frontService{}
	frontSrv.UserHttp = frontHttp.NewUserAppHttpService(
		frontHttp.WithUserDomainService(userDomain),
		frontHttp.WithUserRpcClient(loginClient),
		frontHttp.WithTracer(tracer),
	)
	adminSrv := &adminService{}
	adminSrv.UserGrpcService = adminGrpc.NewUserAppGrpcService()
	// inject cron event of match
	matchEvent := event.NewMatchEvent(
		event.WithTicker(time.NewTicker(time.Second*30)),
		event.WithMatchCronAction(func() error {
			logger.Info("start pull data from baidu")
			return nil
		}),
	)
	App = app{}
	App.FrontService = frontSrv
	App.AdminService = adminSrv
	App.Event = []transport.Server{matchEvent}
	App.Tracer = tracer
}

// InitJaeger 初始化一个opentracing.Tracer链路追踪实例
// 100%的请求都会记录跨度
// 初始化jaeger的指标
func InitJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := &jaegerConfig.Configuration{
		ServiceName: service, // 指定了要被追踪的服务的名称
		Sampler: &jaegerConfig.SamplerConfig{
			// 采用恒定采样策略
			// 意味着对于每一个请求或操作，都会按照固定的方式决定是否进行追踪
			Type: "const",
			// 与 Type 字段配合，决定具体的采样行为
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			// 设置为 true，表示要记录追踪的跨度（Span）信息到日志中
			LogSpans: true,
			// 指定了 Jaeger 收集器（Collector）的端点地址
			// 追踪数据最终需要发送到 Jaeger 的收集器进行处理和存储
			// 通过设置这个字段，告诉程序将追踪数据发送到哪里
			// 指定客户端（应用程序）将追踪数据发送到的目标地址，即 Jaeger 收集器（Collector）的端点
			// TODO 客户端主动push
			//  Jaeger 的这种配置下，是客户端主动将数据推送给收集器，
			// 	这种方式使得客户端对数据的发送有更多的控制权，能够根据自身的情况（如数据量、网络状况等）来决定何时发送数据
			//	 而不是等待服务端来请求
			// TODO 需要做优化更改为kafka - 异步PUSH
			CollectorEndpoint: config.Conf.JaegerConfig.Addr,
		},
	}
	// 基于前面初始化好的配置结构体 cfg
	// 使用 NewTracer 方法来创建一个 Jaeger 追踪器（Tracer）以及一个用于关闭追踪器相关资源的函数 closer
	// jaegerConfig.Logger(jaeger.StdLogger) 是在为追踪器设置日志记录器
	tracer, closer, err := cfg.NewTracer(jaegerConfig.Logger(jaeger.StdLogger))
	if err != nil {
		logger.Fatal(err)
	}
	return tracer, closer
}
