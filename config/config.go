package config

import (
	"encoding/json"
	"os"
)

type Config struct {
	Log struct {
		Level        string   `json:"level"`
		Encoding     string   `json:"encoding"`
		Mode         string   `json:"mode"`
		RedactFields []string `json:"redact_fields"`
	} `json:"log"`
	Mongo struct {
		DatabaseUri  string `json:"database_uri"`
		DatabaseName string `json:"database_name"`
	} `json:"mongo"`
	Grpc struct {
		OrderSvcAddr string `json:"order_svc_addr"`
	} `json:"grpc"`
}

func Load() (Config, error) {
	var cfg Config

	// Set defaults
	cfg.Log.Level = "info"
	cfg.Log.Encoding = "json"
	cfg.Log.Mode = "production"
	cfg.Mongo.DatabaseUri = "mongodb://localhost:27017"
	cfg.Mongo.DatabaseName = "payment"
	cfg.Grpc.OrderSvcAddr = "localhost:50054"

	// Load from environment variables if present
	if mongoUri := os.Getenv("MONGO_URI"); mongoUri != "" {
		cfg.Mongo.DatabaseUri = mongoUri
	}
	if dbName := os.Getenv("MONGO_DB_NAME"); dbName != "" {
		cfg.Mongo.DatabaseName = dbName
	}
	if orderSvcAddr := os.Getenv("ORDER_SVC_ADDR"); orderSvcAddr != "" {
		cfg.Grpc.OrderSvcAddr = orderSvcAddr
	}
	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.Log.Level = logLevel
	}
	if logEncoding := os.Getenv("LOG_ENCODING"); logEncoding != "" {
		cfg.Log.Encoding = logEncoding
	}
	if logMode := os.Getenv("LOG_MODE"); logMode != "" {
		cfg.Log.Mode = logMode
	}

	// Override with config file if it exists
	if _, err := os.Stat("config.json"); err == nil {
		f, err := os.Open("config.json")
		if err != nil {
			return cfg, err
		}
		defer f.Close()

		if err := json.NewDecoder(f).Decode(&cfg); err != nil {
			return cfg, err
		}
	}

	return cfg, nil
}
