package router

import (
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/weiqiangxu/micro_project/common-config/metrics"
	"github.com/weiqiangxu/micro_project/user/application"
	"github.com/weiqiangxu/micro_project/user/config"
	"github.com/weiqiangxu/micro_project/user/global/pprof_tool"
)

func Init(r *gin.Engine) {
	monitorHandle := metrics.PrometheusInterceptor()
	// 注入Prometheus的指标采集的拦截器
	r.Use(monitorHandle)
	// 注入OpenTracing的指标采集用的拦截器
	r.Use(RequestTracingInterceptor())
	// 注册pprof性能分析工具
	pprof_tool.Register(r)
	game := r.Group("/user")
	{
		game.GET("/list", application.App.FrontService.UserHttp.GetUserList)
		game.GET("/info", application.App.FrontService.UserHttp.GetUserInfo)
		game.GET("/detail", application.App.FrontService.UserHttp.GetUserDetail)
	}
}

// RegisterPrometheus 指标收集器注册到Prometheus如果不注册也不会将指标输出到指标采集接口
func RegisterPrometheus() {
	if !config.Conf.HttpConfig.Prometheus {
		return
	}
	prometheus.MustRegister(metrics.RequestLatencyHistogram)
	prometheus.MustRegister(metrics.RequestGauge)
	prometheus.MustRegister(metrics.Counter)
	prometheus.MustRegister(metrics.Summary)
}
