package config

import (
	"strings"

	"github.com/caarlos0/env/v9"
	"github.com/joho/godotenv"
)

type Config struct {
	Log            LogConfig
	PaymentGateway PaymentGatewayConfig
	Grpc           GrpcMicroserviceConfig
	Http           HttpConfig
	Temporal       TemporalConfig
}

type LogConfig struct {
	Level        string   `env:"LOG_LEVEL" envDefault:"debug"`
	Encoding     string   `env:"LOG_ENCODING" envDefault:"development"`
	Mode         string   `env:"LOG_MODE" envDefault:"console"`
	RedactFields []string `env:"LOG_REDACT_FIELDS" envDefault:"password,token,secret"`
}

type PaymentGatewayConfig struct {
	Zalopay ZalopayConfig
}

type ZalopayConfig struct {
	AppID int    `env:"ZALOPAY_APP_ID" envDefault:"1234567890"`
	Key1  string `env:"ZALOPAY_KEY1" envDefault:"1234567890"`
	Key2  string `env:"ZALOPAY_KEY2" envDefault:"1234567890"`
	Host  string `env:"NGROK_TEST_URL" envDefault:""`
}

type TemporalConfig struct {
	HostPort  string `env:"TEMPORAL_HOST_PORT" envDefault:"localhost:7233"`
	Namespace string `env:"TEMPORAL_NAMESPACE" envDefault:"default"`
}

type GrpcMicroserviceConfig struct {
	OrderSvcAddr string `env:"ORDER_SERVICE_ADDRESS" envDefault:"localhost:50054"`
}

type HttpConfig struct {
	Port string `env:"HTTP_PORT" envDefault:"8080"`
}

func Load() (*Config, error) {
	godotenv.Load()
	cfg := &Config{}
	err := env.Parse(cfg)
	if err != nil {
		return nil, err
	}

	// Process the LOG_REDACT_FIELDS env var
	if len(cfg.Log.RedactFields) == 1 && strings.Contains(cfg.Log.RedactFields[0], ",") {
		cfg.Log.RedactFields = strings.Split(cfg.Log.RedactFields[0], ",")
	}

	return cfg, nil
}
