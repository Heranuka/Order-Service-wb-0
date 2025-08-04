package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
	"github.com/joho/godotenv"
)

type Config struct {
	Env      string         `yaml:"env"`
	Http     HTTPConfig     `yaml:"http"`
	Redis    RedisConfig    `yaml:"redis"`
	Postgres PostgresConfig `yaml:"postgres"`
	Kafka    KafkaConfig    `yaml:"kafka"`
}

type HTTPConfig struct {
	Port            string        `yaml:"port"`
	ReadTimeout     time.Duration `yaml:"read_timeout" env-default:"10s"`
	WriteTimeout    time.Duration `yaml:"write_timeout" env-default:"10s"`
	ShutdownTimeout time.Duration `yaml:"shutdown_timeout" env-default:"10s"`
}

type RedisConfig struct {
	Addrs    []string `yaml:"addrs" `
	Password string   `yaml:"password"`
	DBRedis  int      `yaml:"db_redis"`
}

type PostgresConfig struct {
	PostgresURL string `env:"POSTGRES_URL" `
}

type KafkaConfig struct {
	BrokerList                    []string      `yaml:"brokers" `
	Topic                         string        `yaml:"topic" `
	InitialBackoff                time.Duration `yaml:"initial_backoff"`
	MaxRetries                    int           `yaml:"max_retries"`
	TreatUnmarshalErrorAsCritical bool          `yaml:"treatUnmarshalErrorAsCritical"`
	ConsumerGroup                 string        `yaml:"consumer_group" `
}

func LoadConfig() (*Config, error) {
	configPath := fetchConfigPath()

	if configPath == "" {
		return nil, fmt.Errorf("config file is empty")
	}

	return LoadPath(configPath)

}

func LoadPath(configPath string) (*Config, error) {
	err := godotenv.Load()
	if err != nil {
		return nil, fmt.Errorf("error loading .env file")
	}

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return nil, fmt.Errorf("config file does not exist: %v", configPath)
	}
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		return nil, fmt.Errorf("could not read the config: %w", err)
	}
	fmt.Printf("Config: %+v\n", cfg)
	fmt.Printf("Env: %s\n", cfg.Env)
	fmt.Printf("Http: %+v\n", cfg.Http)
	fmt.Printf("Redis: %+v\n", cfg.Redis)
	fmt.Printf("Kafka: %+v\n", cfg.Kafka)
	fmt.Printf("Kafka brokers: %+v\n", cfg.Kafka.BrokerList)

	fmt.Printf("Http Port: %s\n", cfg.Http.Port)
	fmt.Printf("Redis Addrs: %v\n", cfg.Redis.Addrs)
	cfg.Postgres.PostgresURL = os.Getenv("POSTGRES_URL")

	return &cfg, nil
}

func fetchConfigPath() string {
	var configPath string
	flag.StringVar(&configPath, "config", "", "path to configPath")
	flag.Parse()

	if configPath == "" {
		configPath = "local.yaml"
	}
	return configPath
}
