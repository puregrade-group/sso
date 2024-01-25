package config

import (
	"flag"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type (
	// Config -.
	Config struct {
		Env         string        `yaml:"env" env-default:"local"`
		App         AppConfig     `yaml:"app"`
		GRPC        GRPCConfig    `yaml:"grpc"`
		StoragePath string        `yaml:"storage_path"`
		JWTTokenTTL time.Duration `yaml:"jwt_secret_ttl" env-default:"1h"`
	}

	// AppConfig -.
	AppConfig struct {
		Name    string `yaml:"name"`
		Version string `yaml:"version"`
	}

	// GRPCConfig -.
	GRPCConfig struct {
		Port    string        `env-required:"true" yaml:"port"`
		Host    string        `yaml:"host" env-default:""`
		Timeout time.Duration `yaml:"timeout" env-default:"5s"`
	}
)

func MustLoad() *Config {
	configPath := fetchConfigPath()
	if configPath == "" {
		panic("config path is empty")
	}

	return MustLoadPath(configPath)
}

func MustLoadPath(configPath string) *Config {
	// check if file exists
	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		panic("config file does not exist: " + configPath)
	}

	var cfg Config

	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		panic("cannot read config: " + err.Error())
	}

	return &cfg
}

// fetchConfigPath fetches config path from command line flag or environment variable.
// Priority: flag > env > default.
// Default value is empty string.
func fetchConfigPath() string {
	var res string

	flag.StringVar(&res, "config", "", "path to config file")
	flag.Parse()

	if res == "" {
		res = os.Getenv("CONFIG_PATH")
	}

	return res
}
