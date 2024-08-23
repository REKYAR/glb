package main

import (
	"flag"
	"log"
)

func main() {
	configPath := flag.String("config", "", "path to the config file")
	flag.Parse()
	cfgReader := JsonConfigReader{Path: *configPath}
	config, err := cfgReader.ReadConfig()
	if err != nil {
		log.Fatalf("Error reading config: %s", err)
		return
	}
	LoadBalancer := NewLoadBalancer(&config)
	LoadBalancer.Serve()
}
