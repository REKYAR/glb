package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestLoadBalancer(t *testing.T) {
	server1 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			fmt.Printf("/h Server 1 %s\n", r.URL.Host)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Healthy"))
		}
		if r.URL.Path == "/" {
			fmt.Printf("/ Server 1 %s\n", r.URL.Host)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Server 1"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Wrong path, womp womp"))
		}
	}))
	print(server1.URL)
	defer server1.Close()

	server2 := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/health" {
			fmt.Printf("/h Server 2 %s\n", r.URL)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Healthy"))
		}
		if r.URL.Path == "/" {
			fmt.Printf("/ Server 2 %s\n", r.URL.Host)
			w.WriteHeader(http.StatusOK)
			w.Write([]byte("Server 2"))
		} else {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte("Wrong path, womp womp"))
		}
	}))
	defer server2.Close()
	config := &Config{
		Host:                          "localhost",
		Port:                          8080,
		InitialAddresses:              []string{server1.URL, server2.URL},
		Protocol:                      "http",
		HealthCheckPath:               "/health",
		HealthCheckInterval:           1000,
		HealthCheckTimeout:            500,
		HealthCheckUnhealthyThreshold: 200,
		HealthCheckDownInterval:       5000,
	}

	// t.Run("TestInitialHostCheck", func(t *testing.T) {
	// 	lb := NewLoadBalancer(config)
	// 	lb.InitialHostCheck()
	// 	time.Sleep(100 * time.Millisecond)

	// 	for _, host := range config.InitialAddresses {
	// 		status, ok := lb.HostStatus.Load(host)
	// 		if !ok {
	// 			t.Errorf("Host status not set for %s", host)
	// 		}
	// 		if status != HTTP_STATUS_HEALTHY && status != HTTP_STATUS_DOWN && status != HTTP_STATUS_HIGH_LATENCY {
	// 			t.Errorf("Unexpected status for host %s: %s", host, status)
	// 		}
	// 	}
	// })

	t.Run("TestServeHTTP", func(t *testing.T) {
		lb, err := NewLoadBalancer(config)
		if err != nil {
			t.Fatalf("Failed to create LoadBalancer: %v", err)
		}

		// Step 3: Start the LoadBalancer in a separate goroutine
		errCh := make(chan error, 1)

		go func() {
			err := lb.ServeHTTP()
			if err != nil && err != http.ErrServerClosed {
				errCh <- fmt.Errorf("LoadBalancer ServeHTTP failed: %v", err)
			} else {
				errCh <- nil
			}
		}()

		//this hangs the code
		// if err := <-errCh; err != nil {
		// 	t.Fatalf(err.Error())
		// }

		// Give the server a moment to start
		time.Sleep(100 * time.Millisecond)

		// Step 4: Test the ServeHTTP method indirectly through the Serve method
		client := &http.Client{}
		req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
		res, err := client.Do(req)
		if err != nil {
			t.Fatalf("Failed to send request to LoadBalancer: %v", err)
		}
		body, _ := io.ReadAll(res.Body)
		res.Body.Close()
		fmt.Printf("body 1 -%s-", string(body)) // Use fmt.Printf instead of fmt.Fprintf
		if res.StatusCode != http.StatusOK {
			t.Errorf("expected status 200 OK, got %v", res.Status)
		}
		if string(body) != "Server 1" && string(body) != "Server 2" {
			t.Errorf("unexpected response body: %v", string(body))
		}
		// Step 5: Check the load balancing behavior
		// Send multiple requests and ensure they are distributed among the servers
		requestCount := 10
		server1Count := 0
		server2Count := 0

		for i := 0; i < requestCount; i++ {
			fmt.Printf("Request %d\n", i)
			req, _ := http.NewRequest(http.MethodGet, "http://localhost:8080/", nil)
			res, err := client.Do(req)
			if err != nil {
				t.Fatalf("Failed to send request to LoadBalancer: %v", err)
			}
			body, _ := ioutil.ReadAll(res.Body)
			res.Body.Close()

			if string(body) == "Server 1" {
				server1Count++
			} else if string(body) == "Server 2" {
				server2Count++
			}
		}

		if server1Count == 0 || server2Count == 0 {
			t.Errorf("load balancing failed, server1Count: %d, server2Count: %d", server1Count, server2Count)
		}

		// Shutdown the server
		lbServer := &http.Server{
			Addr: "localhost:8080",
		}
		lbServer.Close()
	})

	// 	t.Run("TestUpdateAliveHosts", func(t *testing.T) {
	// 		lb := NewLoadBalancer(config)
	// 		lb.HostStatus.Store(config.InitialAddresses[0], HTTP_STATUS_HEALTHY)
	// 		lb.UpdateAliveHosts()
	// 		time.Sleep(100 * time.Millisecond)

	// 		status, _ := lb.HostStatus.Load(config.InitialAddresses[0])
	// 		if status != HTTP_STATUS_HEALTHY && status != HTTP_STATUS_DOWN && status != HTTP_STATUS_HIGH_LATENCY {
	// 			t.Errorf("Unexpected status after update: %s", status)
	// 		}
	// 	})

	// 	t.Run("TestUpdateDownHosts", func(t *testing.T) {
	// 		lb := NewLoadBalancer(config)
	// 		lb.HostStatus.Store(config.InitialAddresses[0], HTTP_STATUS_DOWN)
	// 		lb.UpdateDownHosts()
	// 		time.Sleep(100 * time.Millisecond)

	// 		status, _ := lb.HostStatus.Load(config.InitialAddresses[0])
	// 		if status != HTTP_STATUS_HEALTHY && status != HTTP_STATUS_DOWN && status != HTTP_STATUS_HIGH_LATENCY {
	// 			t.Errorf("Unexpected status after update: %s", status)
	// 		}
	// 	})

	// 	t.Run("TestTimeGet", func(t *testing.T) {
	// 		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 			time.Sleep(50 * time.Millisecond)
	// 			w.WriteHeader(http.StatusOK)
	// 		}))
	// 		defer server.Close()

	// 		lb := NewLoadBalancer(config)
	// 		res, duration, err := lb.timeGet(server.URL)

	// 		if err != nil {
	// 			t.Errorf("timeGet returned an error: %v", err)
	// 		}
	// 		if res.StatusCode != http.StatusOK {
	// 			t.Errorf("Expected status 200, got %d", res.StatusCode)
	// 		}
	// 		if duration < 50*time.Millisecond {
	// 			t.Errorf("Expected duration >= 50ms, got %v", duration)
	// 		}
	// 	})

	// 	t.Run("TestTimeGetTimeout", func(t *testing.T) {
	// 		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
	// 			time.Sleep(600 * time.Millisecond)
	// 			w.WriteHeader(http.StatusOK)
	// 		}))
	// 		defer server.Close()

	// 		config.HealthCheckTimeout = 500 // ms
	// 		lb := NewLoadBalancer(config)
	// 		_, _, err := lb.timeGet(server.URL)

	// 		if err == nil {
	// 			t.Error("Expected timeout error, got nil")
	// 		}
	// 	})

	// 	t.Run("TestHostStatusTransitions", func(t *testing.T) {
	// 		lb := NewLoadBalancer(config)
	// 		host := "http://example.com"

	// 		transitions := []string{
	// 			HTTP_STATUS_UNKNOWN,
	// 			HTTP_STATUS_HEALTHY,
	// 			HTTP_STATUS_HIGH_LATENCY,
	// 			HTTP_STATUS_DOWN,
	// 			HTTP_STATUS_HEALTHY,
	// 		}

	// 		for i, status := range transitions {
	// 			lb.HostStatus.Store(host, status)
	// 			if currentStatus, _ := lb.HostStatus.Load(host); currentStatus != status {
	// 				t.Errorf("Step %d: Expected status %s, got %v", i, status, currentStatus)
	// 			}
	// 		}
	// 	})

	// 	t.Run("TestConcurrentHostStatusUpdates", func(t *testing.T) {
	// 		lb := NewLoadBalancer(config)
	// 		host := "http://example.com"

	// 		var wg sync.WaitGroup
	// 		for i := 0; i < 100; i++ {
	// 			wg.Add(1)
	// 			go func() {
	// 				defer wg.Done()
	// 				lb.HostStatus.Store(host, HTTP_STATUS_HEALTHY)
	// 				lb.HostStatus.Load(host)
	// 			}()
	// 		}
	// 		wg.Wait()

	// 		if status, _ := lb.HostStatus.Load(host); status != HTTP_STATUS_HEALTHY {
	// 			t.Errorf("Expected final status HEALTHY, got %v", status)
	// 		}
	// 	})

	// 	t.Run("TestConfigValidation", func(t *testing.T) {
	// 		invalidConfigs := []Config{
	// 			{InitialAddresses: []string{}},
	// 			{Protocol: "invalid"},
	// 			{HealthCheckInterval: -1},
	// 			{HealthCheckTimeout: -1},
	// 			{HealthCheckUnhealthyThreshold: -1},
	// 			{HealthCheckDownInterval: -1},
	// 		}

	// 		for _, cfg := range invalidConfigs {
	// 			if err := cfg.ValidateConfig(); err == nil {
	// 				t.Errorf("Expected error for invalid config: %+v", cfg)
	// 			}
	// 		}
	// 	})
	// }

	// func TestJsonConfigReader(t *testing.T) {
	// 	configFile, err := os.CreateTemp("", "config*.json")
	// 	if err != nil {
	// 		t.Fatalf("Failed to create temp file: %v", err)
	// 	}
	// 	defer os.Remove(configFile.Name())

	// 	config := Config{
	// 		Host:             "localhost",
	// 		Port:             8080,
	// 		InitialAddresses: []string{"http://localhost:8081", "http://localhost:8082"},
	// 		Protocol:         "http",
	// 	}

	// 	jsonData, err := json.Marshal(config)
	// 	if err != nil {
	// 		t.Fatalf("Failed to marshal config: %v", err)
	// 	}

	// 	if _, err := configFile.Write(jsonData); err != nil {
	// 		t.Fatalf("Failed to write config: %v", err)
	// 	}
	// 	configFile.Close()

	// 	reader := JsonConfigReader{Path: configFile.Name()}
	// 	readConfig, err := reader.ReadConfig()
	// 	if err != nil {
	// 		t.Fatalf("Failed to read config: %v", err)
	// 	}

	// 	if readConfig.Host != config.Host || readConfig.Port != config.Port || readConfig.Protocol != config.Protocol {
	// 		t.Errorf("Read config does not match written config")
	// 	}
}

// func TestLoadBalancerServe(t *testing.T) {
// 	t.Run("TestServeHTTP", func(t *testing.T) {
// 		config := &Config{
// 			Host:                          "localhost",
// 			Port:                          0,
// 			InitialAddresses:              []string{"http://localhost:8081"},
// 			Protocol:                      "http",
// 			HealthCheckInterval:           1000,
// 			HealthCheckTimeout:            500,
// 			HealthCheckUnhealthyThreshold: 200,
// 			HealthCheckDownInterval:       5000,
// 		}
// 		lb := NewLoadBalancer(config)

// 		// Start the server in a goroutine
// 		go func() {
// 			if err := lb.Serve(); err != nil {
// 				t.Errorf("Serve returned an error: %v", err)
// 			}
// 		}()

// 		// Wait for the server to start
// 		time.Sleep(100 * time.Millisecond)

// 		// TODO: Send a test request to the load balancer
// 		// This requires modifying the Serve() method to return the actual port it's listening on
// 	})

// 	t.Run("TestServeUnsupportedProtocol", func(t *testing.T) {
// 		config := &Config{
// 			Protocol: "unsupported",
// 		}
// 		lb := NewLoadBalancer(config)

// 		err := lb.Serve()
// 		if err == nil {
// 			t.Error("Expected error for unsupported protocol, got nil")
// 		}
// 	})
// }
