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
	tracer, _ := InitJaeger(fmt.Sprintf("%s:%s", config.Conf.Application.Name, config.Conf.Application.Version))

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
		ticker := time.NewTicker(1 * time.Second)
		// 使用for循环来持续接收定时器的触发事件
		for range ticker.C {
			logger.Info("当前GRPC连接状态:", userGrpcConn.GetState().String())
		}

		// TODO 获取userGrpcConn的方式更改为GRPC连接池获取
		// userGrpcConn := grpcPool.Get()
		// TODO grpc连接池获取连接的时候怎么判断这个连接目前是空闲的呢,请求发起到请求完成有事件吗
		// (拦截器的defer是可以接收到完成)
		// TODO grpc连接池获取连接的时候这个连接的顺序是怎么保证的呢

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

// InitJaeger returns an instance of Jaeger Tracer that samples 100% of traces and logs all spans to stdout.
// 初始化jaeger的指标
func InitJaeger(service string) (opentracing.Tracer, io.Closer) {
	cfg := &jaegerConfig.Configuration{
		ServiceName: service,
		Sampler: &jaegerConfig.SamplerConfig{
			Type:  "const",
			Param: 1,
		},
		Reporter: &jaegerConfig.ReporterConfig{
			LogSpans:          true,
			CollectorEndpoint: config.Conf.JaegerConfig.Addr,
		},
	}
	tracer, closer, err := cfg.NewTracer(jaegerConfig.Logger(jaeger.StdLogger))
	if err != nil {
		logger.Fatal(err)
	}
	return tracer, closer
}
