package router

import (
	"github.com/gin-gonic/gin"
	"github.com/weiqiangxu/micro_project/user/application"
	"github.com/weiqiangxu/micro_project/user/global/enum"
)

func RequestTracing() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		span := application.App.Tracer.StartSpan(ctx.FullPath())
		ctx.Set(enum.TraceSpanName, span.Context())
		ctx.Next()
		span.Finish()
	}
}
