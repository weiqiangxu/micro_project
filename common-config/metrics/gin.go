package metrics

import (
	"fmt"
	"time"

	"github.com/weiqiangxu/common-config/format"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ServiceName   = "service"
	RequestPath   = "path"
	RequestMethod = "method"
)

var config format.NacosConfig

// RequestLatencyHistogram 直方图获取请求时长分布
var RequestLatencyHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: config.DataId,
	Name:      fmt.Sprintf("%s_request_duration", config.DataId),
	Help:      "request histogram milli seconds",
	Buckets:   []float64{100, 200, 400, 600, 1000, 1500, 2000, 3000, 5000},
}, []string{ServiceName, RequestPath, RequestMethod})

// RequestGauge 仪表盘获取每个时刻请求响应时长
var RequestGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: config.DataId,
	Name:      fmt.Sprintf("%s_request_gauge", config.DataId),
	Help:      "request gauge for api",
}, []string{ServiceName, RequestPath, RequestMethod})

// Counter 计数器
var Counter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: config.NameSpace,
		Name:      "counter",
		Help:      "",
	},
	[]string{"name"},
)

// RequestMonitor monitor request only effect after register RequestLatencyHistogram && RequestGauge
func RequestMonitor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		ctx.Next()
		latency := time.Since(start)
		labels := prometheus.Labels{}
		labels[ServiceName] = config.DataId
		labels[RequestPath] = ctx.FullPath()
		labels[RequestMethod] = ctx.Request.Method
		RequestGauge.With(labels).Set(float64(latency.Milliseconds()))
		RequestLatencyHistogram.With(labels).Observe(float64(latency.Milliseconds()))
	}
}
