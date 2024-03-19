package config

import (
	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
	"log/slog"
	"os"
	"time"
)

type Config struct {
	Nats     Nats     `yaml:"nats"`
	DbConfig DbConfig `yaml:"dbConfig" env-required:"true"`
	Server   Server   `yaml:"server"`
}

type DbConfig struct {
	DbUser  string `yaml:"user" env-required:"true"`
	DbPass  string `yaml:"password" env-required:"true"`
	DbName  string `yaml:"dbName" env-required:"true"`
	SSLmode string `yaml:"sslmode"`
}

type Server struct {
	Timeout     time.Duration `yaml:"timeout" env-default:"5s"`
	IdleTimeout time.Duration `yaml:"idle_timeout" env-default:"60s"`
	Address     string        `yaml:"address" env-default:"localhost:8080"`
}

type Nats struct {
	IpAddr string `yaml:"ipaddr"`
}

const op = "config.MustLoad: "

// MustLoad -- looks for the config by CONFIG_PATH .env variable and marshals .yaml config to Config. Your project must contain local.env file with CONFIG_PATH variable.
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
