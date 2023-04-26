package opareceiver

import (
	"go.opentelemetry.io/collector/config/confighttp"
)

type Config struct {
	Logs LogsConfig `mapstructure:"logs"`
}

type LogsConfig struct {
	//Secret         string                         `mapstructure:"secret"`
	HTTP *confighttp.HTTPServerSettings `mapstructure:"http"`
	//Attributes     map[string]string              `mapstructure:"attributes"`
	//TimestampField string                         `mapstructure:"timestamp_field"`
}
