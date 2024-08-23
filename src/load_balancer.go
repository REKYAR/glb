package main

import (
	"errors"
	"log"
)

type LoadBalancer struct {
	InitalConfig *Config
	CurrentHosts []string
}

func NewLoadBalancer(config *Config) *LoadBalancer {
	return &LoadBalancer{
		InitalConfig: config,
		CurrentHosts: config.InitialAddresses,
	}
}

func (l *LoadBalancer) Serve() error {
	switch l.InitalConfig.Protocol {
	case "http":
		l.ServeHTTP()
		return nil
	case "rpc":
		l.ServeRPC()
		return nil
	default:
		log.Fatalf("Unsupported protocol: %s", l.InitalConfig.Protocol)
		return errors.New("Unsupported protocol")
	}
}
