package pprof_tool

import (
	netPprof "net/http/pprof"

	"github.com/gin-gonic/gin"
)

const (
	// DefaultPrefix url prefix of pprof
	DefaultPrefix = "/debug/pprof"
)

func getPrefix(prefixOptions ...string) string {
	prefix := DefaultPrefix
	if len(prefixOptions) > 0 {
		prefix = prefixOptions[0]
	}
	return prefix
}

// Register 将来自 net/http/pprof 包的标准处理器与所提供的 gin.Engine 结合起来
// prefixOptions 是可选的
// 如果没有 prefixOptions，则使用默认的路径前缀
// 否则，prefixOptions 中的第一个元素将用作路径前缀。
// 使用 net/http/pprof 类库采集性能指标
func Register(r *gin.Engine, prefixOptions ...string) {
	RouteRegister(&(r.RouterGroup), prefixOptions...)
}

// RouteRegister 注入性能分析（Performance Profiling）的数据查找路由
func RouteRegister(route *gin.RouterGroup, prefixOptions ...string) {
	prefix := getPrefix(prefixOptions...)
	prefixRouter := route.Group(prefix)
	{
		prefixRouter.GET("/", gin.WrapF(netPprof.Index))
		prefixRouter.GET("/cmdline", gin.WrapF(netPprof.Cmdline))
		prefixRouter.GET("/profile", gin.WrapF(netPprof.Profile))
		prefixRouter.POST("/symbol", gin.WrapF(netPprof.Symbol))
		prefixRouter.GET("/symbol", gin.WrapF(netPprof.Symbol))
		prefixRouter.GET("/trace", gin.WrapF(netPprof.Trace))
		prefixRouter.GET("/allocs", gin.WrapH(netPprof.Handler("allocs")))
		prefixRouter.GET("/block", gin.WrapH(netPprof.Handler("block")))
		prefixRouter.GET("/goroutine", gin.WrapH(netPprof.Handler("goroutine")))
		prefixRouter.GET("/heap", gin.WrapH(netPprof.Handler("heap")))
		prefixRouter.GET("/mutex", gin.WrapH(netPprof.Handler("mutex")))
		prefixRouter.GET("/threadcreate", gin.WrapH(netPprof.Handler("threadcreate")))
	}
}
