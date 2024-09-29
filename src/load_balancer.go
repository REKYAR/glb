package main

import (
	"errors"
	"log"
	"net/url"
	"sync"
)

type LoadBalancer struct {
	Config      *Config
	HostStatus  *sync.Map
	HostLatency *sync.Map
	parsedURLs  *sync.Map
	currentIdx  int
}

func NewLoadBalancer(config *Config) (*LoadBalancer, error) {
	status := sync.Map{}
	latency := sync.Map{}
	parsedURLs := sync.Map{}
	for _, host := range config.InitialAddresses {
		status.Store(host, HTTP_STATUS_UNKNOWN)
		latency.Store(host, -1)
		url, error := url.Parse(host)
		if error != nil {
			return nil, errors.New("Error parsing URL: " + error.Error())
		}
		parsedURLs.Store(host, url)
	}
	return &LoadBalancer{
		Config:      config,
		HostStatus:  &status,
		HostLatency: &latency,
		parsedURLs:  &parsedURLs,
		currentIdx:  0,
	}, nil
}

func (l *LoadBalancer) getNextURL() *url.URL {
	for s := 0; s < len(l.Config.InitialAddresses); s++ {
		l.currentIdx = (l.currentIdx + 1) % len(l.Config.InitialAddresses)
		// If all hosts are down, return nil
		status, ok := l.HostStatus.Load(l.Config.InitialAddresses[l.currentIdx])
		if !ok || status == HTTP_STATUS_DOWN {
			continue
		}
		status, ok = l.parsedURLs.Load(l.currentIdx)
		if !ok {
			return nil
		}
		return status.(*url.URL)
	}
	return nil
}

func (l *LoadBalancer) Serve() error {
	if err := l.Config.ValidateConfig(); err != nil {
		return err
	}
	switch l.Config.Protocol {
	case "http":
		l.ServeHTTP()
		return nil
	case "rpc":
		l.ServeRPC()
		return nil
	default:
		log.Fatalf("Unsupported protocol: %s", l.Config.Protocol)
		return errors.New("Unsupported protocol")
	}
}
