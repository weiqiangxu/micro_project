package common

import (
	redisApi "github.com/weiqiangxu/micro_project/common-config/cache"
	"github.com/weiqiangxu/micro_project/common-config/logger"

	"github.com/gin-gonic/gin"
)

// LimitFrequency gin 中间件用于限频
func LimitFrequency(secretKey string, redisApi *redisApi.RedisApi) gin.HandlerFunc {
	return func(c *gin.Context) {
		logger.Info(secretKey)
		logger.Errorf("%+v", redisApi)
		c.Next()
	}
}
