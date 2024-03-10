package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
)

type Config struct {
	DbConfig DbConfig `yaml:"dbConfig" env-required:"true"`
}

type DbConfig struct {
	DbUser  string `yaml:"user" env-required:"true"`
	DbPass  string `yaml:"password" env-required:"true"`
	DbName  string `yaml:"dbName" env-required:"true"`
	SSLmode string `yaml:"sslmode"`
}

const op = "config.MustLoad: "

func MustLoad() *Config {
	if err := godotenv.Load("local.env"); err != nil {
		slog.Error(op, err)
		os.Exit(1)
	}

	configPath := os.Getenv("CONFIG_PATH")
	if configPath == "" {
		slog.Error(op, "config path is empty")
		os.Exit(1)
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		slog.Error(op, "config file doesn't exist", err)
		os.Exit(1)
	}

	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		slog.Error(op, "couldn't read config", err)
		os.Exit(1)
	}

	return &cfg
}
