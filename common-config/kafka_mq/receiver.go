package kafka_mq

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/weiqiangxu/common-config/logger"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

const (
	DefaultTimeOutMs = 6000
)

type ReceiverConfig struct {
	// SASL mechanism to use for authentication. Supported: GSSAPI, PLAIN, SCRAM-SHA-256, SCRAM-SHA-512, OAUTHBEARER. NOTE: Despite the name only one mechanism must be configured.
	Mechanism  string
	UserName   string
	Password   string
	Addr       []string
	GroupName  string
	Topics     []string
	TimeOutMs  int64
	AutoCommit bool
	StdOut     bool
}

type Receiver interface {
	Start(ctx context.Context) error
	Stop(ctx context.Context) error
	SetReceiveChan(ch chan *kafka.Message) error
	Commit(msg *kafka.Message) error
}

func NewReceiver(config *ReceiverConfig) (Receiver, error) {
	if config == nil {
		return nil, errors.New("can't start receiver without config")
	}
	r := &receiver{}
	con := kafka.ConfigMap{}
	err := con.SetKey("bootstrap.servers", strings.Join(config.Addr, ","))
	if err != nil {
		return nil, err
	}
	err = con.SetKey("broker.address.family", "v4")
	if err != nil {
		return nil, err
	} // ip地址类型
	err = con.SetKey("group.id", config.GroupName)
	if err != nil {
		return nil, err
	} // 消费组
	err = con.SetKey("session.timeout.ms", DefaultTimeOutMs)
	if err != nil {
		return nil, err
	} // 超时时间
	err = con.SetKey("auto.offset.reset", "earliest")
	if err != nil {
		return nil, err
	} // 自动offset 规则
	err = con.SetKey("enable.auto.commit", config.AutoCommit)
	if err != nil {
		return nil, err
	} // 是否默认提交，默认为false
	if config.UserName != "" && config.Mechanism != "" {
		err := con.SetKey("sasl.mechanism", config.Mechanism)
		if err != nil {
			return nil, err
		}
		err = con.SetKey("sasl.username", config.UserName)
		if err != nil {
			return nil, err
		} // kafka用户
		err = con.SetKey("sasl.password", config.Password)
		if err != nil {
			return nil, err
		} // kafka密码
	}
	c, err := kafka.NewConsumer(&con)
	if err != nil {
		return nil, err
	}
	r.consumer = c
	r.isStdOut = config.StdOut
	if err := r.consumer.SubscribeTopics(config.Topics, nil); err != nil {
		return nil, err
	}
	return r, nil
}

type receiver struct {
	consumer    *kafka.Consumer
	receiveChan chan *kafka.Message
	isStdOut    bool
}

func (r *receiver) Start(ctx context.Context) error {
	if r.receiveChan == nil {
		logger.Fatal("can't use nil chan to receive kafka_mq msg")
	}
	logger.Info("kafka receiver start!")
	for {
		select {
		case <-ctx.Done():
			logger.Infof("kafka_mq receiver stopping")
			return nil
		default:
			r.consume()
		}
	}
}

// Stop 停止运行
func (r *receiver) Stop(ctx context.Context) error {
	if r.consumer == nil {
		return nil
	}
	r.consumer.Close()
	return nil
}

func (r *receiver) consume() {
	defer func() {
		if r := recover(); r != nil {
			logger.Errorf("consumer recover err=%v", r)
		}
	}()
	msg, err := r.consumer.ReadMessage(100 * time.Millisecond)
	if err != nil {
		time.Sleep(10 * time.Millisecond) // 如果错误阻塞10毫秒，再排队
		return
	}
	if r.isStdOut {
		logger.Infof("handlePush partition =%d data = %s", msg.TopicPartition.Partition, msg.Value)
	}
	r.receiveChan <- msg
}

// SetReceiveChan 设置管道
func (r *receiver) SetReceiveChan(ch chan *kafka.Message) error {
	r.receiveChan = ch
	return nil
}

// Commit 手动确认数据
func (r *receiver) Commit(msg *kafka.Message) error {
	var err error
	if _, err = r.consumer.CommitMessage(msg); err != nil {
		logger.Errorf("CommitMessage partition =%d data = %s err=%v,", msg.TopicPartition.Partition, msg.Value, err)
	}
	return nil
}
