package main

import (
	"github.com/weiqiangxu/micro_project/common-config/format"
	"github.com/weiqiangxu/micro_project/common-config/logger"
	"github.com/weiqiangxu/micro_project/net"
	"github.com/weiqiangxu/micro_project/net/transport"
	"github.com/weiqiangxu/micro_project/net/transport/grpc"
	"github.com/weiqiangxu/micro_project/protocol/user"
	"github.com/weiqiangxu/micro_project/user/application"
	"github.com/weiqiangxu/micro_project/user/config"
	"github.com/weiqiangxu/micro_project/user/global/router"
)

func main() {
	// 配置依赖注入
	config.Conf = config.Config{
		Application:          config.AppInfo{Name: "server", Version: "v0.0.1"},
		UserGrpcServerConfig: format.GrpcConfig{Addr: ":9191"},
	}
	// mongodb && redis 等服务依赖
	application.Init()
	router.RegisterPrometheus()
	// 注入GRPC服务启动时候的监听地址
	grpcServer := grpc.NewServer(grpc.Address(config.Conf.UserGrpcServerConfig.Addr))
	// 将获取用户信息的接口实现注入GRPC服务
	user.RegisterLoginServer(grpcServer, application.App.AdminService.UserGrpcService)
	serverList := []transport.Server{grpcServer}
	if len(application.App.Event) > 0 {
		serverList = append(serverList, application.App.Event...)
	}
	// 将grpc && http 服务注入应用
	app := net.New(
		net.Name(config.Conf.Application.Name),
		net.Version(config.Conf.Application.Version),
		net.Server(serverList...),
	)
	if err := app.Run(); err != nil {
		logger.Fatal(err)
	}
}
