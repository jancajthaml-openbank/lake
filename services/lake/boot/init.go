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

package boot

import (
	"context"
	"os"

	"github.com/jancajthaml-openbank/lake/config"
	"github.com/jancajthaml-openbank/lake/metrics"
	"github.com/jancajthaml-openbank/lake/relay"
	"github.com/jancajthaml-openbank/lake/utils"
)

// Program encapsulate initialized application
type Program struct {
	cfg       config.Configuration
	interrupt chan os.Signal
	metrics   metrics.Metrics
	relay     relay.Relay
	cancel    context.CancelFunc
}

// Initialize application
func Initialize() Program {
	ctx, cancel := context.WithCancel(context.Background())

	cfg := config.GetConfig()

	utils.SetupLogger(cfg.LogLevel)

	metricsDaemon := metrics.NewMetrics(ctx, cfg.MetricsContinuous, cfg.MetricsOutput, cfg.MetricsRefreshRate)
	relayDaemon := relay.NewRelay(ctx, cfg.PullPort, cfg.PubPort, &metricsDaemon)

	return Program{
		cfg:       cfg,
		interrupt: make(chan os.Signal, 1),
		metrics:   metricsDaemon,
		relay:     relayDaemon,
		cancel:    cancel,
	}
}
