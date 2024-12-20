package config

import (
	"github.com/weiqiangxu/micro_project/common-config/format"
)

var Conf Config

type Config struct {
	Application          AppInfo            `toml:"application" json:"application"`
	HttpConfig           format.HttpConfig  `toml:"http_config" json:"http_config"`
	UserGrpcConfig       format.GrpcConfig  `toml:"user_grpc_config" json:"user_grpc_config"`
	UserGrpcServerConfig format.GrpcConfig  `toml:"user_grpc_server_config" json:"user_grpc_server_config"`
	OrderGrpcConfig      format.GrpcConfig  `toml:"order_grpc_config" json:"order_grpc_config"`
	LogConfig            format.LogConfig   `toml:"log_config" json:"log_config"`
	WikiMongoDb          format.MongoConfig `toml:"wiki_mongo_db" json:"wiki_mongo_db"`
	WikiRedisDb          format.RedisConfig `toml:"wiki_redis_db" json:"wiki_redis_db"`
	JwtConfig            JwtConfig          `toml:"jwt_config" json:"jwt_config"`
	JaegerConfig         JaegerConfig       `toml:"jaeger_config" json:"jaeger_config"`
}

type JaegerConfig struct {
	Addr string `toml:"addr"`
}

type JwtConfig struct {
	Secret  string `toml:"secret"`
	Timeout int64  `toml:"timeout"`
}

type AppInfo struct {
	Name    string `toml:"name"`
	Version string `toml:"version"`
}
