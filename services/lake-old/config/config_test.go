package config

import (
	"os"
	"strings"
	"testing"
)

func TestGetConfig(t *testing.T) {
	for _, v := range os.Environ() {
		k := strings.Split(v, "=")[0]
		if strings.HasPrefix(k, "LAKE") {
			os.Unsetenv(k)
		}
	}

	t.Log("has defaults for all values")
	{
		config := LoadConfig()

		if config.PullPort != 5562 {
			t.Errorf("PullPort default value is not 5562")
		}
		if config.PubPort != 5561 {
			t.Errorf("PubPort default value is not 5561")
		}
		if config.LogLevel != "INFO" {
			t.Errorf("LogLevel default value is not INFO")
		}
		if config.MetricsStastdEndpoint != "127.0.0.1:8125" {
			t.Errorf("MetricsStastdEndpoint default value is not 127.0.0.1:8125")
		}
	}
}
