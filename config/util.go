package config

import (
	"context"
	"log"
	"time"

	"github.com/sethvargo/go-envconfig"
)

var configInstance *Config = nil

func GetConfig() *Config {
	if configInstance == nil {
		if err := InitConfig(); err != nil {
			log.Fatalln("Failed to init bot configuration:", err)
		}
	}

	return configInstance
}

func InitConfig() error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	configInstance = &Config{}

	return envconfig.Process(ctx, configInstance)
}
