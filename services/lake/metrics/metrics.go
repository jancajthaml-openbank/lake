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

package metrics

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/jancajthaml-openbank/lake/utils"
	localfs "github.com/jancajthaml-openbank/local-fs"
)

// Metrics holds metrics counters
type Metrics struct {
	utils.DaemonSupport
	storage        localfs.PlaintextStorage
	continuous     bool
	refreshRate    time.Duration
	messageEgress  *uint64
	messageIngress *uint64
}

// NewMetrics returns blank metrics holder
func NewMetrics(ctx context.Context, continuous bool, output string, refreshRate time.Duration) Metrics {
	egress := uint64(0)
	ingress := uint64(0)

	// FIXME can panic
	return Metrics{
		DaemonSupport:  utils.NewDaemonSupport(ctx, "metrics"),
		storage:        localfs.NewPlaintextStorage(output),
		continuous:     continuous,
		refreshRate:    refreshRate,
		messageEgress:  &egress,
		messageIngress: &ingress,
	}
}

// MessageEgress increment number of outcomming messages
func (metrics *Metrics) MessageEgress() {
	atomic.AddUint64(metrics.messageEgress, 1)
}

// MessageIngress increment number of incomming messages
func (metrics *Metrics) MessageIngress() {
	atomic.AddUint64(metrics.messageIngress, 1)
}

// Start handles everything needed to start metrics daemon
func (metrics Metrics) Start() {
	ticker := time.NewTicker(metrics.refreshRate)
	defer ticker.Stop()

	if metrics.continuous {
		metrics.Hydrate()
	}

	metrics.Persist()
	metrics.MarkReady()

	select {
	case <-metrics.CanStart:
		break
	case <-metrics.Done():
		metrics.MarkDone()
		return
	}

	log.Info().Msgf("Start metrics daemon, update each %v into %v", metrics.refreshRate, metrics.storage.Root)

	go func() {
		for {
			select {
			case <-metrics.Done():
				metrics.Persist()
				metrics.MarkDone()
				return
			case <-ticker.C:
				metrics.Persist()
			}
		}
	}()

	metrics.WaitStop()
	log.Info().Msg("Stop metrics daemon")
}
