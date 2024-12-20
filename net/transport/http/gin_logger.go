package http

import (
	"net"
	"net/http"
	"net/http/httputil"
	"os"
	"runtime/debug"
	"strings"
	"time"

	"github.com/weiqiangxu/micro_project/common-config/logger"

	"github.com/gin-gonic/gin"
)

type GinLoggerConfig struct {
	TimeFormat string
	UTC        bool
	SkipPaths  []string
}

// GinZapWithConfig returns a gin.HandlerFunc using configs
func GinZapWithConfig(conf *GinLoggerConfig) gin.HandlerFunc {
	skipPaths := make(map[string]bool, len(conf.SkipPaths))
	for _, path := range conf.SkipPaths {
		skipPaths[path] = true
	}
	return func(c *gin.Context) {
		start := time.Now()
		// some evil middlewares modify this values
		path := c.Request.URL.Path
		query := c.Request.URL.RawQuery
		c.Next()

		if _, ok := skipPaths[path]; !ok {
			end := time.Now()
			latency := end.Sub(start)
			if conf.UTC {
				end = end.UTC()
			}
			messages := []interface{}{
				"method", c.Request.Method,
				"status", c.Writer.Status(),
				"query", query,
				"ip", c.ClientIP(),
				"user-agent", c.Request.UserAgent(),
				"latency", latency,
				"time", end.Format(conf.TimeFormat),
			}
			logger.Infow(path, messages...)
		}
	}
}

// RecoveryWithZap returns a gin.HandlerFunc (middleware)
// that recovers from any panics and logs requests using uber-go/zap.
// All errors are logged using zap.Error().
// stack means whether output the stack info.
// The stack info is easy to find where the error occurs but the stack info is too large.
func RecoveryWithZap(stack bool) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if err := recover(); err != nil {
				// Check for a broken connection, as it is not really a
				// condition that warrants a panic stack trace.
				var brokenPipe bool
				if ne, ok := err.(*net.OpError); ok {
					if se, ok := ne.Err.(*os.SyscallError); ok {
						isBrokenPipe := strings.Contains(strings.ToLower(se.Error()), "broken pipe")
						isReset := strings.Contains(strings.ToLower(se.Error()), "connection reset by peer")
						if isBrokenPipe || isReset {
							brokenPipe = true
						}
					}
				}
				httpRequest, _ := httputil.DumpRequest(c.Request, false)
				if brokenPipe {
					message := []interface{}{
						"error", err,
						"request", string(httpRequest),
					}
					logger.Errorw(c.Request.URL.Path, message...)
					// If the connection is dead, we can't write a status to it.
					e := c.Error(err.(error))
					if e != nil {
						logger.Error(e.Error())
					}
					c.Abort()
					return
				}
				if stack {
					message := []interface{}{
						"time", time.Now(),
						"error", err,
						"request", string(httpRequest),
						"stack", debug.Stack(),
					}
					logger.Errorw("[Recovery from panic]", message...)
				} else {
					message := []interface{}{
						"time", time.Now(),
						"error", err,
						"request", string(httpRequest),
					}
					logger.Errorw("[Recovery from panic]", message...)
				}
				c.AbortWithStatus(http.StatusInternalServerError)
			}
		}()
		c.Next()
	}
}
