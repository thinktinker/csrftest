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

func DefaultPostgresConfig() PostgresConfig {
	return PostgresConfig{
		Host:     "localhost",
		Port:     5432,
		User:     "postgres",
		Password: "",
		Name:     "lenslocked_dev",
	}
}

func (p PostgresConfig) Connection() string {

	if p.Password == "" {
		return fmt.Sprintf("host=%s port=%d user=%s dbname=%s sslmode=disable", p.Host, p.Port, p.User, p.Name)
	}

	return fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=disable", p.Host, p.Port, p.User, p.Password, p.Name)
}

func (p PostgresConfig) Dialect() string {
	return "postgres"
}

type Config struct {
	Port     int            `json:"port"`
	Env      string         `json:"env"`
	Pepper   string         `json:"pepper"`
	HMACKey  string         `json:"hmac_key"`
	Database PostgresConfig `json:"database"`
}

func DefaultConfig() Config {
	return Config{
		Port:     3000,
		Env:      "dev",
		Pepper:   "secret-random-string",
		HMACKey:  "secret-hmac-key",
		Database: DefaultPostgresConfig(),
	}
}

func (c Config) IsProd() bool {
	return c.Env == "prod"
}

func LoadConfig(configReq bool) Config {

	f, err := os.Open(".config")
	if err != nil {
		if configReq {
			panic(err)
		}
		fmt.Println("Using the default config...")
		return DefaultConfig()
	}

	var c Config
	dec := json.NewDecoder(f)
	err = dec.Decode(&c)
	if err != nil {
		panic(err)
	}

	fmt.Println("Successfully loaded config.")
	return c
}
