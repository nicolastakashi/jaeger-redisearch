package model

import (
	"time"

	"github.com/spf13/viper"
)

type Configuration struct {
	MetricsPort       string        `yaml:"metrics_port"`
	RedisAddresses    []string      `yaml:"redis_addresses"`
	RedisWriteTimeout time.Duration `yaml:"redis_write_timeout"`
	RedisTTL          time.Duration `yaml:"redis_ttl"`
}

func InitFromViper(v *viper.Viper) Configuration {
	config := Configuration{}

	v.SetDefault("redis_addresses", []string{"localhost:6379"})
	v.SetDefault("redis_write_timeout", time.Second*30)
	v.SetDefault("redis_ttl", time.Second*60)
	v.SetDefault("metrics_port", "9090")

	config.RedisAddresses = v.GetStringSlice("redis_addresses")
	config.RedisWriteTimeout = v.GetDuration("redis_write_timeout")
	config.RedisTTL = v.GetDuration("redis_ttl")
	config.MetricsPort = v.GetString("metrics_port")

	return config
}
