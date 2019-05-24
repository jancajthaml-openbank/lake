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

package metrics

import (
	"context"
	"fmt"
	"sync/atomic"
	"time"

	"github.com/jancajthaml-openbank/lake/utils"

	log "github.com/sirupsen/logrus"
)

// Metrics holds metrics counters
type Metrics struct {
	utils.DaemonSupport
	output         string
	refreshRate    time.Duration
	messageEgress  *uint64
	messageIngress *uint64
}

// NewMetrics returns blank metrics holder
func NewMetrics(ctx context.Context, output string, refreshRate time.Duration) Metrics {
	egress := uint64(0)
	ingress := uint64(0)

	return Metrics{
		DaemonSupport:  utils.NewDaemonSupport(ctx),
		output:         output,
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

// WaitReady wait for metrics to be ready
func (metrics Metrics) WaitReady(deadline time.Duration) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unknown panic")
			}
		}
	}()

	ticker := time.NewTicker(deadline)
	select {
	case <-metrics.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("daemon was not ready within %v seconds", deadline)
		return
	}
}

// Start handles everything needed to start metrics daemon
func (metrics Metrics) Start() {
	defer metrics.MarkDone()

	ticker := time.NewTicker(metrics.refreshRate)
	defer ticker.Stop()

	if err := metrics.Hydrate(); err != nil {
		log.Warn(err.Error())
	}
	metrics.MarkReady()

	select {
	case <-metrics.CanStart:
		break
	case <-metrics.Done():
		return
	}

	log.Infof("Start metrics daemon, update each %v into %v", metrics.refreshRate, metrics.output)

	for {
		select {
		case <-metrics.Done():
			log.Info("Stopping metrics daemon")
			metrics.Persist()
			log.Info("Stop metrics daemon")
			return
		case <-ticker.C:
			metrics.Persist()
		}
	}
}
