package integration

import (
	"os"
	"testing"

	"github.com/jaegertracing/jaeger/pkg/config"
	"github.com/jaegertracing/jaeger/pkg/metrics"
	"github.com/jaegertracing/jaeger/pkg/testutils"
	"github.com/jaegertracing/jaeger/plugin/storage/grpc"
	"github.com/jaegertracing/jaeger/plugin/storage/integration"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
)

const defaultPluginBinaryPath = "../../../examples/memstore-plugin/memstore-plugin"

type GRPCStorageIntegrationTestSuite struct {
	integration.StorageIntegration
	logger           *zap.Logger
	pluginBinaryPath string
	pluginConfigPath string
}

func (s *GRPCStorageIntegrationTestSuite) initialize() error {
	s.logger, _ = testutils.NewLogger()

	f := grpc.NewFactory()
	v, command := config.Viperize(f.AddFlags)
	flags := []string{
		"--grpc-storage-plugin.binary",
		s.pluginBinaryPath,
		"--grpc-storage-plugin.log-level",
		"debug",
	}
	if s.pluginConfigPath != "" {
		flags = append(flags,
			"--grpc-storage-plugin.configuration-file",
			s.pluginConfigPath,
		)
	}
	err := command.ParseFlags(flags)
	if err != nil {
		return err
	}
	f.InitFromViper(v, zap.NewNop())
	if err = f.Initialize(metrics.NullFactory, s.logger); err != nil {
		return err
	}

	if s.SpanWriter, err = f.CreateSpanWriter(); err != nil {
		return err
	}
	if s.SpanReader, err = f.CreateSpanReader(); err != nil {
		return err
	}

	s.Refresh = s.refresh
	s.CleanUp = s.cleanUp
	return nil
}

func (s *GRPCStorageIntegrationTestSuite) refresh() error {
	return nil
}

func (s *GRPCStorageIntegrationTestSuite) cleanUp() error {
	return s.initialize()
}

func TestGRPCStorage(t *testing.T) {
	if os.Getenv("STORAGE") != "grpc-plugin" {
		t.Skip("Integration test against grpc skipped; set STORAGE env var to grpc-plugin to run this")
	}
	binaryPath := os.Getenv("PLUGIN_BINARY_PATH")
	if binaryPath == "" {
		t.Logf("PLUGIN_BINARY_PATH env var not set, using %s", defaultPluginBinaryPath)
		binaryPath = defaultPluginBinaryPath
	}
	configPath := os.Getenv("PLUGIN_CONFIG_PATH")
	if configPath == "" {
		t.Log("PLUGIN_CONFIG_PATH env var not set")
	}
	s := &GRPCStorageIntegrationTestSuite{
		pluginBinaryPath: binaryPath,
		pluginConfigPath: configPath,
	}
	require.NoError(t, s.initialize())
	s.IntegrationTestAll(t)
}
