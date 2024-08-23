package main

import (
	"io/ioutil"
	"os"
	"testing"
)

func createTempConfigFile(content string) (string, error) {
	tmpfile, err := ioutil.TempFile("", "config-*.json")
	if err != nil {
		return "", err
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		return "", err
	}
	if err := tmpfile.Close(); err != nil {
		return "", err
	}
	return tmpfile.Name(), nil
}

func TestReadConfig(t *testing.T) {
	validConfig := `{
        "Host": "localhost",
        "Port": 8080,
        "InitialAddresses": ["192.168.1.1", "192.168.1.2"],
        "Protocol": "http",
        "HealthCheckInterval": 30,
        "HealthCheckTimeout": 5,
        "HealthCheckUnhealthyThreshold": 3
    }`

	invalidConfig := `{
        "Host": "localhost",
        "Port": "not-an-int",
        "InitialAddresses": ["192.168.1.1", "192.168.1.2"],
        "Protocol": "http",
        "HealthCheckInterval": 30,
        "HealthCheckTimeout": 5,
        "HealthCheckUnhealthyThreshold": 3
    }`

	tests := []struct {
		name          string
		configContent string
		expectError   bool
	}{
		{"ValidConfig", validConfig, false},
		{"InvalidConfig", invalidConfig, true},
		{"NonExistentFile", "", true},
		{"NonJsonFile", "not-a-json-file", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var path string
			var err error

			if tt.name == "NonExistentFile" {
				path = "nonexistent.json"
			} else if tt.name == "NonJsonFile" {
				path = "nonjson.txt"
				err = ioutil.WriteFile(path, []byte(tt.configContent), 0644)
				if err != nil {
					t.Fatalf("Failed to create non-json file: %v", err)
				}
				defer os.Remove(path)
			} else {
				path, err = createTempConfigFile(tt.configContent)
				if err != nil {
					t.Fatalf("Failed to create temp config file: %v", err)
				}
				defer os.Remove(path)
			}

			reader := &JsonConfigReader{Path: path}
			_, err = reader.ReadConfig()
			if (err != nil) != tt.expectError {
				t.Errorf("ReadConfig() error = %v, expectError %v", err, tt.expectError)
			}
		})
	}
}
