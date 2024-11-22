package router

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/weiqiangxu/micro_project/user/application"
	"github.com/weiqiangxu/micro_project/user/global/enum"
)

// RequestTracingInterceptor Http请求的拦截器,在请求进入后
func RequestTracingInterceptor() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		// StartSpan 是 Tracer 对象的一个方法，用于启动一个新的跨度。
		span := application.App.Tracer.StartSpan(fmt.Sprintf("http.request:%s", ctx.FullPath()))
		// 获取当前 Span（跨度）对应的 SpanContext（跨度上下文）
		// SpanContext 则包含了与这个 Span 相关的关键信息，比如该 Span 的唯一标识符、所属的追踪链路、相关的标签等
		spanContext := span.Context()
		// 并且设置到context的集合之中
		ctx.Set(enum.TraceSpanName, spanContext)
		// ctx.Next() 通常表示让请求处理流程继续往下进行
		ctx.Next()
		// Finish() 方法的主要功能是设置当前 Span（跨度）的结束时间戳，并完成对 Span 状态的最终确定
		// 通过 Finish() 方法来标记它的结束，记录下结束的时间点
		span.Finish()
	}
}
