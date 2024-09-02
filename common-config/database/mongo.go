package database

import (
	"context"

	"github.com/weiqiangxu/common-config/format"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
	"go.mongodb.org/mongo-driver/mongo/readpref"
)

func InitMongo(config *format.MongoConfig) (*mongo.Database, error) {
	opt := options.Client()
	opt.Hosts = config.Addr
	if config.User != "" {
		opt.SetAuth(options.Credential{
			AuthSource: config.AuthSource,
			Username:   config.User,
			Password:   config.Passwd,
		})
	}
	// 连接数据库
	client, err := mongo.Connect(context.Background(), opt)
	if err != nil {
		return nil, err
	}
	// 判断服务是不是可用
	if err := client.Ping(context.Background(), readpref.Primary()); err != nil {
		return nil, err
	}
	// 获取数据库和集合
	db := client.Database(config.DB)
	return db, nil
}
