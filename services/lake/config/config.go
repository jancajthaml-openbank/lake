// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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

import "time"

// Configuration of application
type Configuration struct {
	// PullPort represents ZMQ PULL binding
	PullPort int
	// PubPort represents ZMQ PUB binding
	PubPort int
	// LogOutput represents log output
	LogOutput string
	// LogLevel ignorecase log level
	LogLevel string
	// MetricsRefreshRate how frequently should be metrics updated
	MetricsRefreshRate time.Duration
	// MetricsOutput filename of metrics persisted
	MetricsOutput string
}

// Resolver loads config
type Resolver interface {
	GetConfig() Configuration
}

type configResolver struct {
	cfg Configuration
}

// NewResolver provides config resolver
func NewResolver() Resolver {
	return configResolver{cfg: loadConfFromEnv()}
}

// GetConfig loads application configuration
func (c configResolver) GetConfig() Configuration {
	return c.cfg
}
