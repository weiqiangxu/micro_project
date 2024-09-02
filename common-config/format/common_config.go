package format

type NacosConfig struct {
	Host      string `toml:"host" json:"host"`
	Port      uint64 `toml:"port" json:"port"`
	NameSpace string `toml:"name_space" json:"name_space"`
	DataId    string `toml:"data_id" json:"data_id"`
	Group     string `toml:"group" json:"group"`
}

type HttpConfig struct {
	ListenHTTP string `toml:"listen_http" json:"listen_http" validate:"hostname_port" long:"listen_http" description:"[addr]:port for http server to listen"`
	Profile    bool   `toml:"profile" json:"profile" long:"profile" description:"enable go profile"`
	Verbose    bool   `toml:"verbose" json:"verbose" long:"verbose" description:"enable verbose http logging"`
	Tracing    bool   `toml:"tracing" json:"tracing" long:"tracing" description:"enable tracing middleware"`
	Prometheus bool   `toml:"prometheus" json:"prometheus" long:"prometheus" description:"enable prometheus metrics middleware"`
}

type GrpcConfig struct {
	Addr string `toml:"addr" json:"addr" validate:"hostname_port" long:"addr" description:"grpc server addr,format is host:port"`
}

type LogConfig struct {
	Debug  bool   `toml:"debug" json:"debug" long:"debug" description:"enable debug to disable stacktrace"`
	Level  string `toml:"level" json:"level" long:"level" description:"log level, can be empty, or one of debug|info|warn|error|fatal|panic"`
	Output string `toml:"output" json:"output" long:"output" description:"set output file path, can be filepath or stdout|stderr"`
	// Encoding sets the logger's encoding. Valid values are "json" and "console". default: json
	Encoding string `toml:"encoding" json:"encoding" long:"encoding" description:"set encoding, the default is json, if not empty, must be one of: console|json"`
}

type MysqlConfig struct {
	Addr         string `toml:"addr" json:"addr" validate:"hostname_port" long:"addr" description:"mysql server addr,format is host:port"`
	User         string `toml:"user" json:"user" long:"user" description:"mysql user"`
	Passwd       string `toml:"passwd" json:"passwd" long:"passwd" description:"mysql passwd"`
	DB           string `toml:"db" json:"db" validate:"required" long:"db" description:"mysql database name"`
	MaxOpenCount int    `toml:"max_open_count" json:"max_open_count" validate:"required" long:"max_open_count" description:"mysql connection pool max open count"`
	MaxIdleCount int    `toml:"max_idle_count" json:"max_idle_count" validate:"required" long:"max_idle_count" description:"mysql connection pool max idel count"`
	Charset      string `toml:"charset" json:"charset" long:"charset" description:"mysql charset"`
	TimeoutSec   int    `toml:"timeout_sec" json:"timeout_sec" long:"timeout_sec" description:"mysql timeout seconds"` // 超时秒数, 使用时自己拼接DSN &timeout=10s 这样
	Options      string `toml:"options" json:"options" long:"options" description:"mysql extra options, like parseTime=True&loc=Local"`
	Tracing      bool   `toml:"tracing" json:"tracing" long:"tracing" description:"enable tracing middleware"`
}

type RedisConfig struct {
	Addr     string `toml:"addr" json:"addr" validate:"hostname_port" long:"addr" description:"redis server addr,format is host:port"`
	PoolSize int    `toml:"pool_size" json:"pool_size" long:"pool_size" description:"redis connection pool size"`
	Passwd   string `toml:"passwd" json:"passwd" long:"passwd" description:"redis auth passwd, leave it empty if no auth needed"`
}

type MongoConfig struct {
	Addr        []string `toml:"addr" json:"addr" long:"addr" description:"mongo server addr,format is host:port, this option support specific multiple time" validate:"required,dive,hostname_port"`
	User        string   `toml:"user" json:"user" long:"user"`
	Passwd      string   `toml:"passwd" json:"passwd" long:"passwd"`
	AuthSource  string   `toml:"auth_source" json:"auth_source" long:"auth_source"`
	ReplicaSet  string   `toml:"replica_set" json:"replica_set" long:"replica_set"`
	DB          string   `toml:"db" json:"db" long:"db" validate:"required"`
	MinPoolSize uint64   `toml:"min_pool_size" json:"min_pool_size" long:"min_pool_size" validate:"required"`
	// MaxPoolSize The default is 100. If this is 0, it will be set to math.MaxInt64
	MaxPoolSize uint64 `toml:"max_pool_size" json:"max_pool_size" long:"max_pool_size" validate:"required"`
	// MaxConnIdleTime The default is 0, meaning a connection can remain unused indefinitely
	MaxConnIdleTime int64 `toml:"max_conn_idle_time" json:"max_conn_idle_time" long:"max_conn_idle_time"`
	// ConnectTimeout can be set through ApplyURI with the
	// "connectTimeoutMS" (e.g "connectTimeoutMS=30") option. If set to 0, no timeout will be used. The default is 30
	ConnectTimeout int64 `toml:"connect_timeout" json:"connect_timeout" long:"connect_timeout"`
	Tracing        bool  `toml:"tracing" json:"tracing" long:"tracing" description:"enable tracing middleware"`
}

type SentryConfig struct {
	Dsn          string `toml:"dsn" json:"dsn" long:"dsn" description:"sentry DSN url"`
	Env          string `toml:"env" json:"env" validate:"oneof=local dev test prod" long:"env" description:"environment, must be one of local|dev|test|prod"`
	FlushWaitSec int    `toml:"flush_wait_sec" json:"flush_wait_sec" long:"flush_wait_sec" description:"sentry flush wait seconds"` // flush 最多等待秒数
}

type NsqProducerConfig struct {
	Addr string `toml:"addr" json:"addr" long:"addr" validate:"hostname_port"`
}

type NsqConsumerConfig struct {
	Addr        []string `toml:"addr" json:"addr" long:"addr" description:"nsqlookupd addr,format is host:port, this option support specific multiple time" validate:"required,dive,hostname_port"`
	Chan        string   `toml:"chan" json:"chan" long:"chan" validate:"omitempty,ascii"`
	WorkerCount int      `toml:"worker_count" json:"worker_count" long:"worker_count" validate:"required"`
	MaxInFlight int      `toml:"max_in_flight" json:"max_in_flight" long:"max_in_flight" validate:"required"`
}

type KafkaProducerConfig struct {
	Addr []string `toml:"addr" json:"addr" long:"addr" description:"kafka producer addr,format is host:port, this option support specific multiple time" validate:"required,dive,hostname_port"`
}

type KafkaConsumerConfig struct {
	Addr      []string `toml:"addr" json:"addr" long:"addr" description:"kafka consumerGroup addr,format is host:port, this option support specific multiple time" validate:"required,dive,hostname_port"`
	GroupName string   `toml:"group_name" json:"group_name" long:"group_name" description:"kafka consumerGroup group_name" validate:"required"`
}
