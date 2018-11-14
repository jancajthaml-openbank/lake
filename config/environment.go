// Copyright (c) 2016-2018, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"strconv"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const logEnv = "LAKE_LOG"
const logLevelEnv = "LAKE_LOG_LEVEL"
const portPullEnv = "LAKE_PORT_PULL"
const portPubEnv = "LAKE_PORT_PUB"
const metricsRefreshRateEnv = "LAKE_METRICS_REFRESHRATE"
const metricsOutputEnv = "LAKE_METRICS_OUTPUT"

func loadConfFromEnv() Configuration {
	logOutput := getEnvString(logEnv, "")
	metricsOutput := getEnvString(metricsOutputEnv, "")
	metricsRefreshRate := getEnvDuration(metricsRefreshRateEnv, time.Second)
	logLevel := strings.ToUpper(getEnvString(logLevelEnv, "DEBUG"))
	portPub := getEnvInteger(portPubEnv, 5561)
	portPull := getEnvInteger(portPullEnv, 5562)

	return Configuration{
		PullPort:           portPull,
		PubPort:            portPub,
		LogOutput:          logOutput,
		LogLevel:           logLevel,
		MetricsRefreshRate: metricsRefreshRate,
		MetricsOutput:      metricsOutput,
	}
}

func getEnvString(key, fallback string) string {
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
		log.Panicf("invalid value of variable %s", key)
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
		log.Panicf("invalid value of variable %s", key)
	}
	return cast
}