package config

import (
	"fmt"
	"log"

	"github.com/spf13/viper"
)

type Config struct {
	AppHost       string `mapstructure:"APP_HOST"`
	AppPort       int    `mapstructure:"APP_PORT"`
	DatabaseURL   string `mapstructure:"DATABASE_URL"`
	RedisAddr     string `mapstructure:"REDIS_ADDR"`
	RedisPassword string `mapstructure:"REDIS_PASSWORD"`
	RedisDB       int    `mapstructure:"REDIS_DB"`
	LogLevel      string `mapstructure:"LOG_LEVEL"`
}

func Load() (*Config, error) {
	viper.SetConfigName(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")
	viper.AutomaticEnv()

	// Явная привязка переменных окружения
	viper.BindEnv("APP_HOST")
	viper.BindEnv("APP_PORT")
	viper.BindEnv("DATABASE_URL")
	viper.BindEnv("REDIS_ADDR")
	viper.BindEnv("REDIS_PASSWORD")
	viper.BindEnv("REDIS_DB")
	viper.BindEnv("LOG_LEVEL")

	// Defaults
	viper.SetDefault("APP_HOST", "0.0.0.0")
	viper.SetDefault("APP_PORT", 8083)
	viper.SetDefault("REDIS_ADDR", "localhost:6379")
	viper.SetDefault("REDIS_DB", 0)
	viper.SetDefault("LOG_LEVEL", "info")

	if err := viper.ReadInConfig(); err != nil {
		log.Printf("No .env file found, using defaults and env vars: %v", err)
	}

	var cfg Config
	if err := viper.Unmarshal(&cfg); err != nil {
		return nil, err
	}

	if cfg.DatabaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	return &cfg, nil
}
