package model

import (
	"time"

	"github.com/spf13/viper"
)

type Configuration struct {
	RedisAddresses    []string      `yaml:"redis_addresses"`
	RedisWriteTimeout time.Duration `yaml:"redis_write_timeout"`
}

func InitFromViper(v *viper.Viper) Configuration {
	config := Configuration{}

	v.SetDefault("redis_addresses", []string{"localhost:6379"})
	v.SetDefault("redis_write_timeout", time.Second*30)

	config.RedisAddresses = v.GetStringSlice("redis_addresses")
	config.RedisWriteTimeout = v.GetDuration("redis_write_timeout")

	return config
}
