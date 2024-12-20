package grpc_pool

import (
	"fmt"
	"sync"
	"time"

	"github.com/pkg/errors"
	"google.golang.org/grpc"
)

// ErrMaxActiveConnReached 连接池超限
var ErrMaxActiveConnReached = errors.New("MaxActiveConnReached")

// Config 连接池相关配置
type Config struct {
	// 连接池中拥有的最小连接数
	InitialCap int
	// 最大并发存活连接数
	MaxCap int
	// 生成连接的方法
	Factory func() (*grpc.ClientConn, error)
	// 关闭连接的方法
	Close func(*grpc.ClientConn) error
	// 检查连接是否有效的方法
	Ping func(*grpc.ClientConn) error
	// 连接最大空闲时间，超过该事件则将失效
	IdleTimeout time.Duration
}

// channelPool 存放连接信息
type channelPool struct {
	mu                 sync.RWMutex                     // lock
	connections        chan *Conn                       // connection of poll
	factory            func() (*grpc.ClientConn, error) // create connection
	close              func(*grpc.ClientConn) error     // close connection
	ping               func(*grpc.ClientConn) error     // usage to confirm connection us able
	idleTimeout        time.Duration                    // every connection maximum duration available
	maxCap             int                              // maximum number of connections
	openingConnections int                              // open ing connection
}

// Conn connection of grpc
type Conn struct {
	c *grpc.ClientConn
	t time.Time
}

// New 初始化连接
func New(poolConfig *Config) (Pool, error) {
	if poolConfig.InitialCap > poolConfig.MaxCap {
		return nil, errors.New("max must gte init cap")
	}
	if poolConfig.InitialCap == 0 || poolConfig.MaxCap == 0 {
		return nil, errors.New("max or init cap equal zero")
	}
	if poolConfig.Factory == nil {
		return nil, errors.New("invalid factory func settings")
	}
	if poolConfig.Close == nil {
		return nil, errors.New("invalid close func settings")
	}

	c := &channelPool{
		connections:        make(chan *Conn, poolConfig.MaxCap),
		factory:            poolConfig.Factory,
		close:              poolConfig.Close,
		idleTimeout:        poolConfig.IdleTimeout,
		maxCap:             poolConfig.MaxCap,
		openingConnections: poolConfig.InitialCap,
	}

	if poolConfig.Ping != nil {
		c.ping = poolConfig.Ping
	}

	for i := 0; i < poolConfig.InitialCap; i++ {
		connection, err := c.factory()
		if err != nil {
			c.Release()
			return nil, fmt.Errorf("factory is not able to fill the pool: %s", err)
		}
		c.connections <- &Conn{c: connection, t: time.Now()}
	}

	return c, nil
}

// getConnection get all connection in lock
func (c *channelPool) getConnection() chan *Conn {
	c.mu.Lock()
	connections := c.connections
	c.mu.Unlock()
	return connections
}

// Get fetch a connection from pool
func (c *channelPool) Get() (*grpc.ClientConn, error) {
	connections := c.getConnection()
	if connections == nil {
		return nil, ErrClosed
	}
	for {
		select {
		case wrapConn := <-connections:
			if wrapConn == nil {
				// return nil, ErrClosed
				continue
			}
			// whether the timeout occurs, and discard the timeout
			if timeout := c.idleTimeout; timeout > 0 {
				if wrapConn.t.Add(timeout).Before(time.Now()) {
					// close connect once timeout
					_ = c.Close(wrapConn.c)
					continue
				}
			}
			// whether it is invalid. If it is invalid, discard it
			// If the user does not set the ping method, do not check
			// 如果设置了ping方法在调用之前检查一下这个连接
			if c.ping != nil {
				if err := c.Ping(wrapConn.c); err != nil {
					_ = c.Close(wrapConn.c)
					continue
				}
			}
			// append used conn to the end
			c.connections <- wrapConn
			return wrapConn.c, nil
		default:
			c.mu.Lock()
			if c.factory == nil {
				c.mu.Unlock()
				return nil, ErrClosed
			}
			conn, err := c.factory()
			if err != nil {
				c.mu.Unlock()
				return nil, err
			}
			c.openingConnections++
			c.mu.Unlock()
			// add to buff
			c.connections <- &Conn{c: conn, t: time.Now()}
			return conn, nil
		}
	}
}

// Put add the connection back in the pool
func (c *channelPool) Put(conn *grpc.ClientConn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.connections == nil {
		_ = c.Close(conn)
		return errors.New("connection is nil")
	}
	if len(c.connections) >= c.maxCap {
		_ = c.Close(conn)
		return nil
	}
	c.connections <- &Conn{c: conn, t: time.Now()}
	return nil
}

// Close connection close
func (c *channelPool) Close(conn *grpc.ClientConn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	if c.close == nil {
		return nil
	}
	c.openingConnections--
	return c.close(conn)
}

// Ping check availability
func (c *channelPool) Ping(conn *grpc.ClientConn) error {
	if conn == nil {
		return errors.New("connection is nil. rejecting")
	}
	return c.ping(conn)
}

// Release all connections in the connection pool drop
func (c *channelPool) Release() {
	c.mu.Lock()
	connections := c.connections
	c.connections = nil
	c.factory = nil
	c.ping = nil
	c.close = nil
	c.mu.Unlock()
	if connections == nil {
		return
	}
	close(connections)
	for wrapConn := range connections {
		_ = c.close(wrapConn.c)
	}
}

// Len gen connection length
func (c *channelPool) Len() int {
	return len(c.getConnection())
}
