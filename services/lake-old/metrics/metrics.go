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

package metrics

import (
	"runtime"
	"sync/atomic"

	"github.com/DataDog/datadog-go/statsd"
)

// Metrics provides helper function for metrics
type Metrics interface {
	MessageEgress()
	MessageIngress()
}

// StatsdMetrics provides metrics helper with statsd client
type StatsdMetrics struct {
	client         *statsd.Client
	messageEgress  int64
	messageIngress int64
}

// NewMetrics returns blank metrics holder
func NewMetrics(endpoint string) *StatsdMetrics {
	client, err := statsd.New(endpoint, statsd.WithClientSideAggregation(), statsd.WithoutTelemetry())
	if err != nil {
		log.Error().Err(err).Msg("Failed to ensure statsd client")
		return nil
	}
	return &StatsdMetrics{
		client:         client,
		messageEgress:  int64(0),
		messageIngress: int64(0),
	}
}

// MessageEgress increment number of outcomming messages
func (instance *StatsdMetrics) MessageEgress() {
	if instance == nil {
		return
	}
	atomic.AddInt64(&(instance.messageEgress), 1)
}

// MessageIngress increment number of incomming messages
func (instance *StatsdMetrics) MessageIngress() {
	if instance == nil {
		return
	}
	atomic.AddInt64(&(instance.messageIngress), 1)
}

// Setup does nothing
func (*StatsdMetrics) Setup() error {
	return nil
}

// Done returns always finished
func (*StatsdMetrics) Done() <-chan interface{} {
	done := make(chan interface{})
	close(done)
	return done
}

// Cancel triggers work once
func (instance *StatsdMetrics) Cancel() {
	instance.Work()
	instance.client.Flush()
}

// Work represents metrics worker work
func (instance *StatsdMetrics) Work() {
	if instance == nil {
		return
	}

	egress := instance.messageEgress
	ingress := instance.messageIngress

	atomic.AddInt64(&(instance.messageEgress), -egress)
	atomic.AddInt64(&(instance.messageIngress), -ingress)

	var stats = new(runtime.MemStats)
	runtime.ReadMemStats(stats)

	instance.client.Count("openbank.lake.message.ingress", ingress, nil, 1)
	instance.client.Count("openbank.lake.message.egress", egress, nil, 1)
	instance.client.Gauge("openbank.lake.memory.bytes", float64(stats.Sys), nil, 1)
}
