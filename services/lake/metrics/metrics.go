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
	"runtime"
	"sync/atomic"
	"time"

	"github.com/jancajthaml-openbank/lake/support/concurrent"
	localfs "github.com/jancajthaml-openbank/local-fs"
)

// Metrics holds metrics counters
type Metrics struct {
	concurrent.DaemonSupport
	storage         localfs.Storage
	continuous      bool
	refreshRate     time.Duration
	messageEgress   uint64
	messageIngress  uint64
	memoryAllocated uint64
}

// NewMetrics returns blank metrics holder
func NewMetrics(ctx context.Context, continuous bool, output string, refreshRate time.Duration) *Metrics {
	storage, err := localfs.NewPlaintextStorage(output)
	if err != nil {
		log.Error().Msgf("Failed to ensure storage %+v", err)
		return nil
	}
	return &Metrics{
		DaemonSupport:   concurrent.NewDaemonSupport(ctx, "metrics"),
		storage:         storage,
		continuous:      continuous,
		refreshRate:     refreshRate,
		messageEgress:   uint64(0),
		messageIngress:  uint64(0),
		memoryAllocated: uint64(0),
	}
}

// MessageEgress increment number of outcomming messages
func (metrics *Metrics) MessageEgress() {
	if metrics == nil {
		return
	}
	atomic.AddUint64(&(metrics.messageEgress), 1)
}

// MessageIngress increment number of incomming messages
func (metrics *Metrics) MessageIngress() {
	if metrics == nil {
		return
	}
	atomic.AddUint64(&(metrics.messageIngress), 1)
}

// MemoryAllocatedSnapshot updates memory allocated snapshot
func (metrics *Metrics) MemoryAllocatedSnapshot() {
	if metrics == nil {
		return
	}
	var stats = new(runtime.MemStats)
	runtime.ReadMemStats(stats)
	atomic.StoreUint64(&(metrics.memoryAllocated), stats.Sys)
}

// Start handles everything needed to start metrics daemon
func (metrics *Metrics) Start() {
	if metrics == nil {
		return
	}
	ticker := time.NewTicker(metrics.refreshRate)
	defer ticker.Stop()

	if metrics.continuous {
		metrics.Hydrate()
	}

	metrics.MemoryAllocatedSnapshot()
	metrics.Persist()
	metrics.MarkReady()

	select {
	case <-metrics.CanStart:
		break
	case <-metrics.Done():
		metrics.MarkDone()
		return
	}

	log.Info().Msgf("Start metrics daemon, update file each %v", metrics.refreshRate)

	go func() {
		for {
			select {
			case <-metrics.Done():
				metrics.MemoryAllocatedSnapshot()
				metrics.Persist()
				metrics.MarkDone()
				return
			case <-ticker.C:
				metrics.MemoryAllocatedSnapshot()
				metrics.Persist()
			}
		}
	}()

	metrics.WaitStop()
	log.Info().Msg("Stop metrics daemon")
}
