package config

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type Config struct {
	Environment string
	Redis       RedisConfig
	Storage     StorageConfig
	Queues      QueueConfig
	Logging     LoggingConfig
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
	Stream   string
	Group    string
	Consumer string
}

type StorageConfig struct {
	Endpoint        string
	AccessKey       string
	SecretKey       string
	BucketOriginals string
	BucketVariants  string
	UseSSL          bool
	Region          string
}

type QueueConfig struct {
	VisibilityTimeout time.Duration
	ClaimInterval     time.Duration
}

type LoggingConfig struct {
	Level string
}

func Load() (*Config, error) {
	v := viper.New()
	v.SetConfigName("worker")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("../../config")
	v.SetEnvPrefix("NODEIMAGE_WORKER")
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("load config: %w", err)
		}
	}

	var cfg Config
	if err := v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "mapstructure"
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
		)
	}); err != nil {
		return nil, fmt.Errorf("unmarshal: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("environment", "development")
	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.db", 0)
	v.SetDefault("redis.stream", "media:ingest")
	v.SetDefault("redis.group", "media-workers")
	v.SetDefault("redis.consumer", "worker-1")

	v.SetDefault("queues.visibilitytimeout", "2m")
	v.SetDefault("queues.claiminterval", "10s")

	v.SetDefault("logging.level", "info")
}
