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

package metrics

import (
	"encoding/json"
	"os"
	"sync"
	"time"

	"github.com/jancajthaml-openbank/lake/pkg/utils"

	gom "github.com/rcrowley/go-metrics"
	log "github.com/sirupsen/logrus"
)

// Snapshot holds metrics snapshot status
type Snapshot struct {
	MessageEgress  int64 `json:"messageEgress"`
	MessageIngress int64 `json:"messageIngress"`
}

// Metrics holds metrics counters
type Metrics struct {
	messageEgress  gom.Counter
	messageIngress gom.Counter
}

// NewMetrics returns blank metrics holder
func NewMetrics() *Metrics {
	return &Metrics{
		messageEgress:  gom.NewCounter(),
		messageIngress: gom.NewCounter(),
	}
}

// NewSnapshot returns metrics snapshot
func NewSnapshot(entity *Metrics) Snapshot {
	if entity == nil {
		return Snapshot{}
	}

	return Snapshot{
		MessageEgress:  entity.messageEgress.Count(),
		MessageIngress: entity.messageIngress.Count(),
	}
}

// MessageEgress increment number of outcomming messages
func (gom *Metrics) MessageEgress(num int64) {
	gom.messageEgress.Inc(num)
}

// MessageIngress increment number of incomming messages
func (gom *Metrics) MessageIngress(num int64) {
	gom.messageIngress.Inc(num)
}

func (gom *Metrics) persist(filename string) {
	tempFile := filename + "_temp"
	data, err := json.Marshal(NewSnapshot(gom))
	if err != nil {
		log.Warnf("unable to create serialize metrics with error: %v", err)
		return
	}
	f, err := os.OpenFile(tempFile, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, os.ModePerm)
	if err != nil {
		log.Warnf("unable to create file with error: %v", err)
		return
	}
	defer f.Close()

	if _, err := f.Write(data); err != nil {
		log.Warnf("unable to write file with error: %v", err)
		return
	}

	if err := os.Rename(tempFile, filename); err != nil {
		log.Warnf("unable to move file with error: %v", err)
		return
	}

	log.Debugf("metrics updated at %s", filename)
	return
}

// PersistPeriodically stores metrics holded in memory periodically to disk
func PersistPeriodically(wg *sync.WaitGroup, terminationChan chan struct{}, params utils.RunParams, data *Metrics) {
	defer wg.Done()

	if params.MetricsOutput == "" {
		log.Warnf("no metrics output defined, skipping metrics persistence")
		return
	}

	ticker := time.NewTicker(params.MetricsRefreshRate)
	defer ticker.Stop()

	log.Debugf("Updating metrics each %v into %v", params.MetricsRefreshRate, params.MetricsOutput)

	for {
		select {
		case <-ticker.C:
			data.persist(params.MetricsOutput)
		case <-terminationChan:
			data.persist(params.MetricsOutput)
			return
		}
	}
}
