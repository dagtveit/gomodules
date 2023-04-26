package opareceiver

import (
	"context"
	"go.opentelemetry.io/collector/component"
	"go.opentelemetry.io/collector/config/confighttp"
	"go.opentelemetry.io/collector/consumer"
	"go.opentelemetry.io/collector/receiver"
)

const (
	typeStr        = "opa"
	stabilityLevel = component.StabilityLevelAlpha
)

func NewFactory() receiver.Factory {
	return receiver.NewFactory(
		typeStr,
		createDefaultConfig,
		receiver.WithLogs(createLogsReceiver, stabilityLevel),
	)
}

func createDefaultConfig() component.Config {
	return &Config{
		Logs: LogsConfig{
			HTTP: &confighttp.HTTPServerSettings{},
		},
	}
}

func createLogsReceiver(
	ctx context.Context,
	params receiver.CreateSettings,
	rConf component.Config,
	consumer consumer.Logs,
) (receiver.Logs, error) {
	cfg := rConf.(*Config)
	return newLogsReceiver(params, cfg, consumer)
}
