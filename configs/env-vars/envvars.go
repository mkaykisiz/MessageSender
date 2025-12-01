package envvars

import (
	"fmt"
	"time"

	"github.com/codingconcepts/env"
)

// EnvVars represents environment variables
type EnvVars struct {
	Service       Service
	Mongo         Mongo
	Redis         Redis
	Localization  Localization
	HTTPServer    HTTPServer
	Configs       Configs
	MessageClient MessageClient
}

// Configs represents environment configs
type Configs struct {
	StartMessageCount int           `env:"CONFIG_START_MESSAGE_COUNT" default:"2"`
	SendMessageDelay  time.Duration `env:"CONFIG_SEND_MESSAGE_DURATION" default:"120s"`
}

// MessageClient represents message client webhook
type MessageClient struct {
	Url string `env:"MESSAGE_CLIENT_URL" required:"true"`
	AuthKey string `env:"MESSAGE_CLIENT_AUTH_KEY" required:"true"`
	Timeout time.Duration `env:"MESSAGE_CLIENT_TIMEOUT" default:"30s"`
	MaxRetries int `env:"MESSAGE_CLIENT_MAX_RETRIES" default:"3"`
	RetryDelay time.Duration `env:"MESSAGE_CLIENT_RETRY_DELAY" default:"1s"`
}

// Service represents service configurations
type Service struct {
	ProjectName           string        `env:"SERVICE_PROJECT_NAME" required:"true"`
	Name                  string        `env:"SERVICE_NAME" required:"true"`
	Environment           string        `env:"SERVICE_ENVIRONMENT" default:"dev"`
	Release               string        `env:"SERVICE_RELEASE"`
	ShutdownSleepDuration time.Duration `env:"SERVICE_SHUTDOWN_SLEEP_DURATION" default:"30s"`
}

// Mongo represents mongo configurations
type Mongo struct {
	URI               string        `env:"MONGO_URI" required:"true"`
	Database          string        `env:"MONGO_DATABASE" required:"true"`
	ConnectTimeout    time.Duration `env:"MONGO_CONNECT_TIMEOUT" default:"10s"`
	PingTimeout       time.Duration `env:"MONGO_PING_TIMEOUT" default:"10s"`
	ReadTimeout       time.Duration `env:"MONGO_READ_TIMEOUT" default:"10s"`
	WriteTimeout      time.Duration `env:"MONGO_WRITE_TIMEOUT" default:"5s"`
	DisconnectTimeout time.Duration `env:"MONGO_DISCONNECT_TIMEOUT" default:"5s"`
}

// Redis represents redis configurations
type Redis struct {
	Address            string        `env:"REDIS_ADDRESS" required:"true"`
	Username           string        `env:"REDIS_USERNAME"`
	Password           string        `env:"REDIS_PASSWORD"`
	DB                 int           `env:"REDIS_DB"`
	DialTimeout        time.Duration `env:"REDIS_DIAL_TIMEOUT" default:"5s"`
	ReadTimeout        time.Duration `env:"REDIS_READ_TIMEOUT" default:"10s"`
	WriteTimeout       time.Duration `env:"REDIS_WRITE_TIMEOUT" default:"10s"`
	PoolSize           int           `env:"REDIS_POOL_SIZE" default:"10"`
	MinIdleConnections int           `env:"REDIS_MIN_IDLE_CONNECTIONS" default:"5"`
	MaxConnectionAge   time.Duration `env:"REDIS_MAX_CONNECTION_AGE" default:"5m"`
	IdleTimeout        time.Duration `env:"REDIS_IDLE_TIMEOUT" default:"5m"`
}

// Localization represents localization configurations
type Localization struct {
	LanguageFilesDirectory string `env:"LOCALIZATION_LANGUAGE_FILES_DIRECTORY" required:"true"`
}

// HTTPServer represents http server configurations
type HTTPServer struct {
	Address         string        `env:"HTTP_SERVER_ADDRESS" default:":8004"`
	ReadTimeout     time.Duration `env:"HTTP_SERVER_READ_TIMEOUT" default:"15s"`
	WriteTimeout    time.Duration `env:"HTTP_SERVER_WRITE_TIMEOUT" default:"15s"`
	IdleTimeout     time.Duration `env:"HTTP_SERVER_IDLE_TIMEOUT" default:"15s"`
	MaxHeaderBytes  int           `env:"HTTP_SERVER_MAX_HEADER_BYTES" default:"1048576"`
	ShutdownTimeout time.Duration `env:"HTTP_SERVER_SHUTDOWN_TIMEOUT" default:"10s"`
}

// LoadEnvVars loads and returns environment variables
func LoadEnvVars() (*EnvVars, error) {
	s := Service{}
	if err := env.Set(&s); err != nil {
		return nil, fmt.Errorf("loading service environment variables failed, %s", err.Error())
	}

	m := Mongo{}
	if err := env.Set(&m); err != nil {
		return nil, fmt.Errorf("loading mongo environment variables failed, %s", err.Error())
	}

	r := Redis{}
	if err := env.Set(&r); err != nil {
		return nil, fmt.Errorf("loading redis environment variables failed, %s", err.Error())
	}

	l := Localization{}
	if err := env.Set(&l); err != nil {
		return nil, fmt.Errorf("loading localization environment variables failed, %s", err.Error())
	}

	hs := HTTPServer{}
	if err := env.Set(&hs); err != nil {
		return nil, fmt.Errorf("loading http server environment variables failed, %s", err.Error())
	}

	cfg := Configs{}
	if err := env.Set(&cfg); err != nil {
		return nil, fmt.Errorf("loading http server environment variables failed, %s", err.Error())
	}

	msg := MessageClient{}
	if err := env.Set(&msg); err != nil {
		return nil, fmt.Errorf("loading http server environment variables failed, %s", err.Error())
	}

	ev := &EnvVars{
		Service:       s,
		Mongo:         m,
		Redis:         r,
		Localization:  l,
		HTTPServer:    hs,
		Configs:       cfg,
		MessageClient: msg,
	}

	return ev, nil
}
