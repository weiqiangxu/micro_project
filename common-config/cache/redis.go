package redisapi

import (
	"context"
	"errors"
	"math"
	"time"

	"github.com/weiqiangxu/common-config/format"
	"github.com/weiqiangxu/common-config/logger"

	"github.com/gomodule/redigo/redis"
)

// RedisInterface redis interface for all service
type RedisInterface interface {
	Get(key string) (value string, err error)
	SetNxEx(key, value string, expireTs int64) (err error)
	Del(key string) (err error)
	Expire(key string, seconds uint64) (err error)
	GenerateLock(ctx context.Context, SeizeTimeOut time.Duration, uuid string, key string) error // 分布式锁
	ReleaseLock(key string, uuid string)
	MonitorLock(ctx context.Context, key string, expire int64)
}

const (
	DftMaxRedisPoolLimit = 1000
	healthCheckPeriod    = time.Second * 30
	maxDelayTime         = 3
	lockExpireTime       = 10 * time.Second
)

var (
	ErrRedisExecFailed  = errors.New("redis exec failed")
	ErrRedisKeyNotExist = errors.New("redis key does not exist")
	ErrSeizeTimeOut     = errors.New("seize time out")
)

type HandleListLoopMessageFunc func(message *string) (code uint32, err error)

// RedisApi Conn exposes a set of callbacks for the various events that occur on a connection
type RedisApi struct {
	redisPool   *redis.Pool
	redisServer string
}

// NewRedisApi create new *RedisApi with maxPoolSize pool size, AUTH is enabled if passwd is not empty string
func NewRedisApi(redisConfig format.RedisConfig) RedisInterface {
	pool := newRedisPoolWithSizeAndPasswd(redisConfig.Addr, redisConfig.PoolSize, redisConfig.Passwd)
	return &RedisApi{
		redisPool:   pool,
		redisServer: redisConfig.Addr,
	}
}

func newRedisPoolWithSizeAndPasswd(redisServer string, maxPoolSize int, passwd string) *redis.Pool {
	poolSize := DftMaxRedisPoolLimit
	if maxPoolSize != 0 {
		poolSize = maxPoolSize
	}
	return &redis.Pool{
		// Maximum number of idle connections in the pool.
		MaxIdle: 20,
		// Maximum number of connections allocated by the pool at a given time.
		// When zero, there is no limit on the number of connections in the pool
		MaxActive: poolSize,
		// Close connections after remaining idle for this duration. If the value
		// is zero, then idle connections are not closed. Applications should set
		// the timeout to a value less than the server's timeout.
		IdleTimeout: 300 * time.Second,
		// If Wait is true and the pool is at the MaxActive limit, then Get() waits
		// for a connection to be returned to the pool before returning.
		Wait: false,
		// Dial is an application supplied function for creating and configuring a
		// connection.
		//
		// The connection returned from Dial must not be in a special state
		// (subscribed to pubs ub channel, transaction started, ...).
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", redisServer,
				redis.DialConnectTimeout(3*time.Second),
				// Read timeout on server should be greater than ping period.
				redis.DialReadTimeout(healthCheckPeriod+10*time.Second),
				redis.DialWriteTimeout(10*time.Second))
			if err != nil {
				return nil, err
			}
			// 密码非空才认证
			if passwd != "" {
				if _, err := c.Do("AUTH", passwd); err != nil {
					err := c.Close()
					if err != nil {
						return nil, err
					}
					return nil, err
				}
			}
			return c, err
		},
		// TestOnBorrow is an optional application supplied function for checking
		// the health of an idle connection before the connection is used again by
		// the application. Argument t is the time that the connection was returned
		// to the pool. If the function returns an error, then the connection is
		// closed.
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < healthCheckPeriod {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}
}

func (api *RedisApi) Get(key string) (string, error) {
	// 获取一条Redis连接
	redisConn := api.redisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			logger.Errorf("redis get catch err=%#v", err)
		}
	}(redisConn)
	r, err := redisConn.Do("GET", key)
	return redis.String(r, err)
}

// SetNxEx set nx
func (api *RedisApi) SetNxEx(key, value string, expireTs int64) (err error) {
	// 获取一条Redis连接
	redisConn := api.redisPool.Get()
	retValue, err := redis.String(redisConn.Do("SET", key, value, "NX", "EX", expireTs))
	defer func() {
		_ = redisConn.Close()
	}()
	if err != nil {
		return err
	}
	if retValue == "OK" { // 执行成功
		return nil
	}
	return ErrRedisExecFailed
}

// Del del
func (api *RedisApi) Del(key string) (err error) {
	// 获取一条Redis连接
	redisConn := api.redisPool.Get()
	delCount, err := redis.Int(redisConn.Do("DEL", key))
	if err == nil && delCount == 1 {
		err = nil
	} else if err == nil && delCount == 0 {
		err = nil
	}
	err = redisConn.Close()
	if err != nil {
		return err
	}
	return err
}

// Expire 设置Key的生存时间
func (api *RedisApi) Expire(key string, seconds uint64) error {
	// 获取一条Redis连接
	redisConn := api.redisPool.Get()
	defer func(redisConn redis.Conn) {
		err := redisConn.Close()
		if err != nil {
			logger.Errorf("redis expire catch %#v", err)
		}
	}(redisConn)
	retValue, err := redis.Int(redisConn.Do("EXPIRE", key, seconds))
	if err != nil {
		return err
	}
	var valueCheckErr error
	if err == nil {
		switch retValue {
		case 1:
			valueCheckErr = nil
		case 0:
			valueCheckErr = ErrRedisKeyNotExist
		default:
			valueCheckErr = ErrRedisExecFailed
		}
	}
	return valueCheckErr
}

// GenerateLock 生成锁附加自旋
func (api *RedisApi) GenerateLock(ctx context.Context, SeizeTimeOut time.Duration, uuid string, key string) error {
	var seizeCtx context.Context
	if SeizeTimeOut > 0 {
		var cf context.CancelFunc
		seizeCtx, cf = context.WithTimeout(ctx, SeizeTimeOut)
		defer cf()
	}
	for {
		select {
		case <-seizeCtx.Done():
			return ErrSeizeTimeOut
		default:
			err := api.SetNxEx(key, uuid, int64(lockExpireTime.Seconds()))
			if err == nil {
				go api.MonitorLock(ctx, key, int64(lockExpireTime.Seconds()))
				logger.Infof("get key:%v,uuid:%v", key, uuid)
				return nil
			} else {
				logger.Infof("can't get key")
				// 如果当前moment被阻塞 则自旋200毫秒等待，让P去处理其他G
				time.Sleep(200 * time.Millisecond)
			}
		}
	}
}

// ReleaseLock 释放锁
func (api *RedisApi) ReleaseLock(key string, uuid string) {
	if val, _ := api.Get(key); val == uuid {
		_ = api.Del(key)
	}
}

// MonitorLock 检查协程存在情况下需要延期
func (api *RedisApi) MonitorLock(ctx context.Context, key string, expire int64) {
	t := time.Duration(math.Ceil(float64(expire) / 2))
	times := 0
	// 超过自旋次数放弃，数据丢失
	for times < maxDelayTime {
		select {
		case <-ctx.Done():
			return
		default:
			time.Sleep(t) // 休眠
			_ = api.Expire(key, uint64(t))
			times++
		}
	}
}
