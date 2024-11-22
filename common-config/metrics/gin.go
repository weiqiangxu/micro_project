package metrics

import (
	"fmt"
	"time"

	"github.com/weiqiangxu/micro_project/common-config/format"

	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus"
)

const (
	ServiceName   = "service"
	RequestPath   = "path"
	RequestMethod = "method"
)

var config format.NacosConfig

// RequestLatencyHistogram 1.直方图获取请求时长分布(按照范围桶聚集指标)
var RequestLatencyHistogram = prometheus.NewHistogramVec(prometheus.HistogramOpts{
	Namespace: config.DataId,
	Name:      fmt.Sprintf("%s_request_duration", config.DataId),
	Help:      "请求监控的直方图",
	// 设置了桶（buckets）等相关参数用于划分请求延迟时间的不同范围
	// 按照第一个桶0-100ms\第二个桶100-200ms
	Buckets: []float64{100, 200, 400, 600, 1000, 1500, 2000, 3000, 5000},
}, []string{ServiceName, RequestPath, RequestMethod})

// RequestGauge 2.仪表盘获取每个时刻请求响应时长
var RequestGauge = prometheus.NewGaugeVec(prometheus.GaugeOpts{
	Namespace: config.DataId,
	Name:      fmt.Sprintf("%s_request_gauge", config.DataId),
	Help:      "请求响应时长的仪表盘指标",
}, []string{ServiceName, RequestPath, RequestMethod})

// Counter 3.计数器
var Counter = prometheus.NewCounterVec(
	prometheus.CounterOpts{
		Namespace: config.NameSpace,
		Name:      "counter",
		Help:      "计数器",
	},
	[]string{ServiceName, RequestPath, RequestMethod},
)

// Summary 4.记录请求的摘要(关注数据的分位数（如中位数、90 分位数、99 分位数等）)
var Summary = prometheus.NewSummary(
	prometheus.SummaryOpts{
		Name: "summary",
		Help: "记录请求时长分布所在的百分位",
		Objectives: map[float64]float64{
			0.5: 0.05,
			// 0.9这个分位数，它表示 90% 分位数
			// 即收集的数据（例如请求时长）中有 90% 的数据点小于或等于这个 90% 分位数对应的数值
			// 0.01表示该分位数对应的允许误差（tolerance）
			// 具体来说，计算得到的 90% 分位数的实际值与真实的 90% 分位数之间的差异应该在这个误差范围内
			0.9:  0.01,
			0.99: 0.001},
		ConstLabels: map[string]string{}, // 可以添加常量标签，这里假设暂时没有
	})

// PrometheusInterceptor prometheus指标采集用的拦截器
func PrometheusInterceptor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		start := time.Now()
		// 走完请求调用逻辑
		ctx.Next()
		// 记录延迟
		latency := time.Since(start)
		labels := prometheus.Labels{}
		labels[ServiceName] = config.DataId
		labels[RequestPath] = ctx.FullPath()
		labels[RequestMethod] = ctx.Request.Method
		// 请求时长仪表盘记录请求的毫秒数
		RequestGauge.With(labels).Set(float64(latency.Milliseconds()))
		// 创建直方图并且打标签
		// 注入延迟的毫秒
		RequestLatencyHistogram.With(labels).Observe(float64(latency.Milliseconds()))
		// 计数器
		Counter.With(labels).Inc()
		// 注入请求时长记录百分位
		Summary.Observe(float64(latency.Milliseconds()))
	}
}
