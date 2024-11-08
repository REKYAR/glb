package main

import (
	"encoding/json"
	"errors"
	"log"
	"os"
)

type ConfigReader interface {
	ReadConfig() (Config, error)
}

type Config struct {
	Host                          string
	Port                          int
	InitialAddresses              []string
	Protocol                      string
	HealthCheckPath               string
	HealthCheckInterval           int //ms
	HealthCheckTimeout            int //ms
	HealthCheckUnhealthyThreshold int //ms
	HealthCheckDownInterval       int //ms
}

type JsonConfigReader struct {
	Path string
}

func (j *JsonConfigReader) ReadConfig() (Config, error) {
	_, err := os.Stat(j.Path)
	if err != nil {
		log.Printf("Config file not found: %s", j.Path)
		return Config{}, err
	}
	if len(j.Path) < 5 && j.Path[len(j.Path)-5:] != ".json" {
		log.Printf("Config file must be a JSON file")
		return Config{}, err
	}
	file, err := os.Open(j.Path)
	if err != nil {
		log.Printf("Error opening config file: %s", j.Path)
		return Config{}, err
	}
	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Printf("Error decoding config file: %s", j.Path)
		return Config{}, err
	}
	return config, nil
}

func (c *Config) ValidateConfig() error {
	if len(c.InitialAddresses) == 0 {
		return errors.New("InitialAddresses cannot be empty")
	}
	if c.Protocol != "http" && c.Protocol != "rpc" {
		return errors.New("Unsupported protocol")
	}
	if c.HealthCheckInterval <= 0 {
		return errors.New("HealthCheckInterval must be positive")
	}
	if c.HealthCheckTimeout <= 0 {
		return errors.New("HealthCheckTimeout must be positive")
	}
	if c.HealthCheckUnhealthyThreshold <= 0 {
		return errors.New("HealthCheckUnhealthyThreshold must be positive")
	}
	if c.HealthCheckDownInterval <= 0 {
		return errors.New("HealthCheckDownInterval must be positive")
	}
	return nil
}
