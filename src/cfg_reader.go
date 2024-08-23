package main

import (
	"encoding/json"
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
	HealthCheckInterval           int
	HealthCheckTimeout            int
	HealthCheckUnhealthyThreshold int
}

type JsonConfigReader struct {
	Path string
}

func (j *JsonConfigReader) ReadConfig() (Config, error) {
	_, err := os.Stat(j.Path)
	if err != nil {
		log.Fatalf("Config file not found: %s", j.Path)
		return Config{}, err
	}
	if len(j.Path) < 0 && j.Path[len(j.Path)-5:] != ".json" {
		log.Fatalf("Config file must be a JSON file")
		return Config{}, err
	}
	file, err := os.Open(j.Path)
	if err != nil {
		log.Fatalf("Error opening config file: %s", j.Path)
		return Config{}, err
	}
	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Error decoding config file: %s", j.Path)
		return Config{}, err
	}
	return config, nil
}
