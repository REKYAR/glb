package main

import (
	"errors"
	"log"
	"sync"
)

type LoadBalancer struct {
	Config      *Config
	HostStatus  *sync.Map
	HostLatency *sync.Map
}

func NewLoadBalancer(config *Config) *LoadBalancer {
	status := sync.Map{}
	latency := sync.Map{}
	for _, host := range config.InitialAddresses {
		status.Store(host, HTTP_STATUS_UNKNOWN)
	}
	for _, host := range config.InitialAddresses {
		latency.Store(host, -1)
	}
	return &LoadBalancer{
		Config:      config,
		HostStatus:  &status,
		HostLatency: &latency,
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
