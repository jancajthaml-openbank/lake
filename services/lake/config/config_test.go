package config

import (
	"os"
	"strings"
	"testing"
	"time"
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
		if config.MetricsContinuous != true {
			t.Errorf("MetricsContinuous default value is not true")
		}
		if config.MetricsRefreshRate != time.Second {
			t.Errorf("MetricsRefreshRate default value is not 1s")
		}
		if config.MetricsOutput != "/tmp/lake-metrics" {
			t.Errorf("MetricsOutput default value is not /tmp/lake-metrics")
		}
	}
}
