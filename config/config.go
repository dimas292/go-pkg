package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	App AppConfig `yaml:"app"`
}

type AppConfig struct {
	Name string    `yaml:"name"`
	Port string    `yaml:"port"`
	Jwt  JWTConfig `yaml:"jwt"`
	Db   DbConfig  `yaml:"db"`
}

type JWTConfig struct {
	Secret     string `yaml:"secret"`
	Expiration int    `yaml:"expiration"` // hours
}

type DbConfig struct {
	Postgres PostgresConfig `yaml:"postgres"`
	Redis    RedisConfig    `yaml:"redis"`
}

type PostgresConfig struct {
	DBHost     string `yaml:"dbhost"`
	DBUser     string `yaml:"dbuser"`
	DBPassword string `yaml:"dbpassword"`
	DBName     string `yaml:"dbname"`
}

func (c PostgresConfig) DSN() string {
	return fmt.Sprintf(
		"host=%s user=%s password=%s dbname=%s sslmode=disable",
		c.DBHost, c.DBUser, c.DBPassword, c.DBName,
	)
}

type RedisConfig struct {
	Host string `yaml:"host"`
	Port string `yaml:"port"`
}

func (c RedisConfig) Addr() string {
	return fmt.Sprintf("%s:%s", c.Host, c.Port)
}

func LoadConfig(path string) (*Config, error) {
	f, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	cfg := &Config{}
	err = yaml.Unmarshal(f, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	return cfg, nil
}
