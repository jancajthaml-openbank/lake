// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"github.com/jancajthaml-openbank/lake/env"
)

// Configuration of application
type Configuration struct {
	// PullPort represents ZMQ PULL binding
	PullPort int
	// PubPort represents ZMQ PUB binding
	PubPort int
	// LogLevel ignorecase log level
	LogLevel string
	// MetricsStastdEndpoint represents statsd daemon hostname
	MetricsStastdEndpoint string
}

// LoadConfig loads application configuration
func LoadConfig() Configuration {
	return Configuration{
		PullPort:              env.Int("LAKE_PORT_PULL", 5562),
		PubPort:               env.Int("LAKE_PORT_PUB", 5561),
		LogLevel:              strings.ToUpper(env.String("LAKE_LOG_LEVEL", "INFO")),
		MetricsStastdEndpoint: env.String("LAKE_STATSD_ENDPOINT", "127.0.0.1:8125"),
	}
}
