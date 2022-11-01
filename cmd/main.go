package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"strings"

	"github.com/nicolastakashi/jaeger-redisearch/internal/model"
	"github.com/nicolastakashi/jaeger-redisearch/internal/repository"
	"github.com/nicolastakashi/jaeger-redisearch/internal/store"
	"github.com/prometheus/client_golang/prometheus/promhttp"

	_ "net/http/pprof"

	"github.com/hashicorp/go-hclog"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc/shared"
	"github.com/jaegertracing/jaeger/storage/dependencystore"
	"github.com/jaegertracing/jaeger/storage/spanstore"
	"github.com/rueian/rueidis"
	"github.com/spf13/viper"
)

var configPath string

func main() {
	flag.StringVar(&configPath, "config", "", "A path to the plugin's configuration file")
	flag.Parse()

	logger := hclog.New(&hclog.LoggerOptions{
		Name:       "jaeger-redisearch",
		Level:      hclog.Warn, // Jaeger only captures >= Warn, so don't bother logging below Warn
		JSONFormat: true,
	})

	v := viper.New()
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(strings.NewReplacer("-", "_", ".", "_"))

	if configPath != "" {
		v.SetConfigFile(configPath)
		err := v.ReadInConfig()
		if err != nil {
			logger.Error("failed to parse configuration file", "err", err)
			os.Exit(1)
		}
	}

	config := model.InitFromViper(v)

	c, err := rueidis.NewClient(rueidis.ClientOption{
		InitAddress:      config.RedisAddresses,
		ConnWriteTimeout: config.RedisWriteTimeout,
		ClientName:       "jaeger-redisearch",
	})

	if err != nil {
		logger.Error("error to connect to redis", err)
		os.Exit(1)
	}

	defer c.Close()

	spanRepository, err := repository.NewSpanRepository(logger, c, config)

	if err != nil {
		logger.Error("error to create span repository", err)
		os.Exit(1)
	}

	serviceRepository, err := repository.NewOperationRepository(logger, c, config)

	if err != nil {
		logger.Error("error to create span repository", err)
		os.Exit(1)
	}

	go func() {
		http.Handle("/metrics", promhttp.Handler())
		// http.Handle("/debug/pprof/heap", _)
		err = http.ListenAndServe(fmt.Sprintf(":%v", config.HttpPort), nil)
		if err != nil {
			logger.Error("Failed to listen for metrics endpoint", "error", err)
		}
	}()

	plugin := &RedisStorePlugin{
		writer: store.NewSpanWriter(logger, spanRepository, serviceRepository),
		reader: store.NewSpanReader(logger, spanRepository, serviceRepository),
	}

	grpc.Serve(&shared.PluginServices{
		Store: plugin,
	})
}

type RedisStorePlugin struct {
	reader *store.SpanReader
	writer *store.SpanWriter
}

func (s *RedisStorePlugin) DependencyReader() dependencystore.Reader {
	return s.reader
}

func (s *RedisStorePlugin) SpanReader() spanstore.Reader {
	return s.reader
}

func (s *RedisStorePlugin) SpanWriter() spanstore.Writer {
	return s.writer
}
