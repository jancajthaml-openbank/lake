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
	"bytes"
	"context"
	"fmt"
	"strconv"
	"sync/atomic"
	"time"

	"github.com/jancajthaml-openbank/lake/utils"
)

// Metrics holds metrics counters
type Metrics struct {
	utils.DaemonSupport
	output         string
	continuous     bool
	refreshRate    time.Duration
	messageEgress  *uint64
	messageIngress *uint64
}

// NewMetrics returns blank metrics holder
func NewMetrics(ctx context.Context, continuous bool, output string, refreshRate time.Duration) Metrics {
	egress := uint64(0)
	ingress := uint64(0)

	return Metrics{
		DaemonSupport:  utils.NewDaemonSupport(ctx),
		output:         output,
		continuous:     continuous,
		refreshRate:    refreshRate,
		messageEgress:  &egress,
		messageIngress: &ingress,
	}
}

// MarshalJSON serialises Metrics as json bytes
func (metrics *Metrics) MarshalJSON() ([]byte, error) {
	if metrics == nil {
		return nil, fmt.Errorf("cannot marshall nil")
	}

	if metrics.messageEgress == nil || metrics.messageIngress == nil {
		return nil, fmt.Errorf("cannot marshall nil references")
	}

	var buffer bytes.Buffer

	buffer.WriteString("{\"messageEgress\":")
	buffer.WriteString(strconv.FormatUint(*metrics.messageEgress, 10))
	buffer.WriteString(",\"messageIngress\":")
	buffer.WriteString(strconv.FormatUint(*metrics.messageIngress, 10))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// UnmarshalJSON deserializes Metrics from json bytes
func (metrics *Metrics) UnmarshalJSON(data []byte) error {
	if metrics == nil {
		return fmt.Errorf("cannot unmarshall to nil")
	}

	if metrics.messageEgress == nil || metrics.messageIngress == nil {
		return fmt.Errorf("cannot unmarshall to nil references")
	}

	aux := &struct {
		MessageEgress  uint64 `json:"messageEgress"`
		MessageIngress uint64 `json:"messageIngress"`
	}{}

	if err := utils.JSON.Unmarshal(data, &aux); err != nil {
		return err
	}

	atomic.StoreUint64(metrics.messageEgress, aux.MessageEgress)
	atomic.StoreUint64(metrics.messageIngress, aux.MessageIngress)

	return nil
}
