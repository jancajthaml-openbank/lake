// Copyright (c) 2016-2020, Jan Cajthaml <jan.cajthaml@gmail.com>
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package config

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"
	log "github.com/sirupsen/logrus"
)

func loadConfFromEnv() Configuration {
	metricsOutput := getEnvFilename("LAKE_METRICS_OUTPUT", "/tmp")
	metricsContinuous := getEnvBoolean("LAKE_METRICS_CONTINUOUS", true)
	metricsRefreshRate := getEnvDuration("LAKE_METRICS_REFRESHRATE", time.Second)
	logLevel := strings.ToUpper(getEnvString("LAKE_LOG_LEVEL", "INFO"))
	portPub := getEnvInteger("LAKE_PORT_PUB", 5561)
	portPull := getEnvInteger("LAKE_PORT_PULL", 5562)

	if metricsOutput != "" && os.MkdirAll(filepath.Dir(metricsOutput), os.ModePerm) != nil {
		log.Fatal("unable to assert metrics output")
	}

	return Configuration{
		PullPort:           portPull,
		PubPort:            portPub,
		LogLevel:           logLevel,
		MetricsContinuous:  metricsContinuous,
		MetricsRefreshRate: metricsRefreshRate,
		MetricsOutput:      metricsOutput,
	}
}

func getEnvBoolean(key string, fallback bool) bool {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := strconv.ParseBool(value)
	if err != nil {
		log.Errorf("invalid value of variable %s", key)
		return fallback
	}
	return cast
}

func getEnvFilename(key string, fallback string) string {
	var value = strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	value = filepath.Clean(value)
	if os.MkdirAll(value, os.ModePerm) != nil {
		return fallback
	}
	return value
}

func getEnvString(key string, fallback string) string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	return value
}

func getEnvInteger(key string, fallback int) int {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := strconv.Atoi(value)
	if err != nil {
		log.Errorf("invalid value of variable %s", key)
		return fallback
	}
	return cast
}

func getEnvDuration(key string, fallback time.Duration) time.Duration {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return fallback
	}
	cast, err := time.ParseDuration(value)
	if err != nil {
		log.Errorf("invalid value of variable %s", key)
		return fallback
	}
	return cast
}
