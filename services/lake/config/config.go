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
	"strings"
	"time"
)

// Configuration of application
type Configuration struct {
	// PullPort represents ZMQ PULL binding
	PullPort int
	// PubPort represents ZMQ PUB binding
	PubPort int
	// LogLevel ignorecase log level
	LogLevel string
	// MetricsContinuous determines if metrics should start from last state
	MetricsContinuous bool
	// MetricsRefreshRate how frequently should metrics be updated
	MetricsRefreshRate time.Duration
	// MetricsOutput determines into which filename should metrics write
	MetricsOutput string
}

// GetConfig loads application configuration
func GetConfig() Configuration {
	return Configuration{
		PullPort:           envInteger("LAKE_PORT_PULL", 5562),
		PubPort:            envInteger("LAKE_PORT_PUB", 5561),
		LogLevel:           strings.ToUpper(envString("LAKE_LOG_LEVEL", "INFO")),
		MetricsContinuous:  envBoolean("LAKE_METRICS_CONTINUOUS", true),
		MetricsRefreshRate: envDuration("LAKE_METRICS_REFRESHRATE", time.Second),
		MetricsOutput:      envFilename("LAKE_METRICS_OUTPUT", "/tmp/lake-metrics"),
	}
}
