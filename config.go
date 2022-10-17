package main

import (
	"encoding/json"
	"fmt"
	"os"
)

type PostgresConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	Name     string `json:"name"`
}

func (c PostgresConfig) Dialect() string {
	return "postgres"
}

func (c PostgresConfig) ConnectionString() string {
	if c.Password == "" {
		return fmt.Sprintf("host=%s port=%v user=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Name)
	}
	return fmt.Sprintf("host=%s port=%v user=%s password=%s dbname=%s sslmode=disable", c.Host, c.Port, c.User, c.Password, c.Name)

}
func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "admin",
		Password: "qwerty",
		Name:     "photogallery_dev",
	}
}

type Config struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACkey  string         `json:"hamc_key"`
	Database PostgresConfig `json:"database"`
}

func (c *Config) IsProd() bool {
	return c.Env == "production" || c.Env == "Prod" || c.Env == "prod"
}

func DefaultConfig() Config {
	return Config{
		Port:     3000,
		Env:      "dev",
		Pepper:   "secret-random-string-dev",
		HMACkey:  "secret-hmac-key-dev",
		Database: DefaultPostgresConfig(),
	}
}

func LoadConfig(isProd bool) Config {
	f, err := os.Open(".config")
	if err != nil {
		if isProd {
			panic(err)
		}
		fmt.Println("Using the default config.")
		return DefaultConfig()
	}
	var c Config
	dec := json.NewDecoder(f)
	err = dec.Decode(&c)
	must(err)
	fmt.Println("Successfully loaded .config")
	return c
}
