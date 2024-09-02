package kafka_mq

import (
	"errors"
	"strings"

	"github.com/weiqiangxu/common-config/logger"

	"github.com/confluentinc/confluent-kafka-go/kafka"
)

type PusherConfig struct {
	Addr      []string
	Mechanism string
	UserName  string
	Password  string
}

type Pusher interface {
	SendMessageWithKey(topic string, key string, message []byte) error
	SendMessage(topic string, message []byte) error
}

func NewPusher(config *PusherConfig) (Pusher, error) {
	if config == nil {
		return nil, errors.New("can't start receiver without config")
	}
	c := kafka.ConfigMap{}
	err := c.SetKey("bootstrap.servers", strings.Join(config.Addr, ","))
	if err != nil {
		return nil, err
	}
	err = c.SetKey("broker.address.family", "v4")
	if err != nil {
		return nil, err
	}
	if config.UserName != "" && config.Mechanism != "" {
		err := c.SetKey("sasl.mechanism", config.Mechanism)
		if err != nil {
			return nil, err
		}
		err = c.SetKey("sasl.username", config.UserName)
		if err != nil {
			return nil, err
		} // kafka用户
		err = c.SetKey("sasl.password", config.Password)
		if err != nil {
			return nil, err
		} // kafka密码
	}
	producer, err := kafka.NewProducer(&c)
	if err != nil {
		logger.Fatal(err)
		return nil, err
	}
	p := &pusher{}
	p.producer = producer
	return p, nil
}

type pusher struct {
	producer *kafka.Producer
}

func (p *pusher) SendMessage(topic string, message []byte) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Value:          message,
	}
	if err := p.producer.Produce(msg, nil); err != nil {
		logger.Errorf("kafka_mq send message err:", err)
	}
	return nil
}

func (p *pusher) SendMessageWithKey(topic string, key string, message []byte) error {
	msg := &kafka.Message{
		TopicPartition: kafka.TopicPartition{Topic: &topic, Partition: kafka.PartitionAny},
		Key:            []byte(key),
		Value:          message,
	}
	if err := p.producer.Produce(msg, nil); err != nil {
		logger.Errorf("kafka_mq send message err:", err)
	}
	return nil
}
