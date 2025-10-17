package config

import (
	"fmt"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/viper"
)

type TLSConfig struct {
	Enabled  bool
	CertFile string
	KeyFile  string
}

type HTTPConfig struct {
	Host         string
	Port         int
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
	IdleTimeout  time.Duration
}

type PostgresConfig struct {
	DSN             string
	MaxOpen         int
	MaxIdle         int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Addr     string
	Password string
	DB       int
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

type SecurityConfig struct {
	JWTAccessSecret  string
	JWTRefreshSecret string
	JWTAccessTTL     time.Duration
	JWTRefreshTTL    time.Duration
	SignatureSecret  string
	MaxSessions      int
}

type NSFWConfig struct {
	ModelPath        string
	ThresholdBlock   float64
	ThresholdReview  float64
	RecheckInterval  time.Duration
}

type AppConfig struct {
	Environment   string
	HTTP          HTTPConfig
	TLS           TLSConfig
	Postgres      PostgresConfig
	Redis         RedisConfig
	Storage       StorageConfig
	Security      SecurityConfig
	NSFW          NSFWConfig
	AllowCORSOrigins []string
}

func Load() (*AppConfig, error) {
	v := viper.New()
	v.SetConfigName("config")
	v.SetConfigType("yaml")
	v.AddConfigPath(".")
	v.AddConfigPath("./config")
	v.AddConfigPath("../config")

	v.SetEnvPrefix("NODEIMAGE")
	v.AutomaticEnv()

	setDefaults(v)

	if err := v.ReadInConfig(); err != nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
			return nil, fmt.Errorf("load config file: %w", err)
		}
	}

	var cfg AppConfig
	if err := v.Unmarshal(&cfg, func(dc *mapstructure.DecoderConfig) {
		dc.TagName = "mapstructure"
		dc.DecodeHook = mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		)
	}); err != nil {
		return nil, fmt.Errorf("unmarshal config: %w", err)
	}

	return &cfg, nil
}

func setDefaults(v *viper.Viper) {
	v.SetDefault("environment", "development")

	v.SetDefault("http.host", "0.0.0.0")
	v.SetDefault("http.port", 8080)
	v.SetDefault("http.readtimeout", "10s")
	v.SetDefault("http.writetimeout", "15s")
	v.SetDefault("http.idletimeout", "60s")

	v.SetDefault("postgres.maxopen", 30)
	v.SetDefault("postgres.maxidle", 10)
	v.SetDefault("postgres.connmaxlifetime", "30m")

	v.SetDefault("redis.addr", "127.0.0.1:6379")
	v.SetDefault("redis.db", 0)

	v.SetDefault("storage.bucketoriginals", "nodeimage-originals")
	v.SetDefault("storage.bucketvariants", "nodeimage-variants")
	v.SetDefault("storage.usessl", false)
	v.SetDefault("storage.region", "us-east-1")

	v.SetDefault("security.jwtaccessttl", "15m")
	v.SetDefault("security.jwtrefreshttl", "720h") // 30 days
	v.SetDefault("security.maxsessions", 10)

	v.SetDefault("nsfw.thresholdblock", 0.92)
	v.SetDefault("nsfw.thresholdreview", 0.75)
	v.SetDefault("nsfw.recheckinterval", "168h")

}
