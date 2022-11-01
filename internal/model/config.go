package model

import (
	"time"

	"github.com/spf13/viper"
)

type Configuration struct {
	HttpPort          string        `yaml:"http_port"`
	RedisAddresses    []string      `yaml:"redis_addresses"`
	RedisWriteTimeout time.Duration `yaml:"redis_write_timeout"`
	RedisTTL          time.Duration `yaml:"redis_ttl"`
}

func InitFromViper(v *viper.Viper) Configuration {
	config := Configuration{}

	v.SetDefault("redis_addresses", []string{"localhost:6379"})
	v.SetDefault("redis_write_timeout", time.Second*30)
	v.SetDefault("redis_ttl", time.Second*60)
	v.SetDefault("http_port", "9090")

	config.RedisAddresses = v.GetStringSlice("redis_addresses")
	config.RedisWriteTimeout = v.GetDuration("redis_write_timeout")
	config.RedisTTL = v.GetDuration("redis_ttl")
	config.HttpPort = v.GetString("http_port")

	return config
}
