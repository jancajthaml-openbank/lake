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
	"bytes"
	"fmt"
	"encoding/json"
	"os"
	"runtime"
	"strconv"
	"sync/atomic"
)

// MarshalJSON serializes Metrics as json bytes
func (metrics *Metrics) MarshalJSON() ([]byte, error) {
	if metrics == nil {
		return nil, fmt.Errorf("cannot marshal nil")
	}

	if metrics.messageEgress == nil || metrics.messageIngress == nil {
		return nil, fmt.Errorf("cannot marshal nil references")
	}

	var stats = new(runtime.MemStats)
	runtime.ReadMemStats(stats)

	var buffer bytes.Buffer

	buffer.WriteString("{\"messageEgress\":")
	buffer.WriteString(strconv.FormatUint(*metrics.messageEgress, 10))
	buffer.WriteString(",\"messageIngress\":")
	buffer.WriteString(strconv.FormatUint(*metrics.messageIngress, 10))
	buffer.WriteString(",\"memoryAllocated\":")
	buffer.WriteString(strconv.FormatUint(stats.Sys, 10))
	buffer.WriteString("}")

	return buffer.Bytes(), nil
}

// UnmarshalJSON deserializes Metrics from json bytes
func (metrics *Metrics) UnmarshalJSON(data []byte) error {
	if metrics == nil {
		return fmt.Errorf("cannot unmarshal to nil")
	}

	if metrics.messageEgress == nil || metrics.messageIngress == nil {
		return fmt.Errorf("cannot unmarshal to nil references")
	}

	aux := &struct {
		MessageEgress  uint64 `json:"messageEgress"`
		MessageIngress uint64 `json:"messageIngress"`
	}{}

	if err := json.Unmarshal(data, &aux); err != nil {
		return err
	}

	atomic.StoreUint64(metrics.messageEgress, aux.MessageEgress)
	atomic.StoreUint64(metrics.messageIngress, aux.MessageIngress)

	return nil
}

// Persist saved metrics state to storage
func (metrics *Metrics) Persist() error {
	if metrics == nil {
		return fmt.Errorf("cannot persist nil reference")
	}
	data, err := json.Marshal(metrics)
	if err != nil {
		return err
	}
	err = metrics.storage.WriteFile("metrics.json", data)
	if err != nil {
		return err
	}
	err = os.Chmod(metrics.storage.Root+"/metrics.json", 0644)
	if err != nil {
		return err
	}
	return nil
}

// Hydrate loads metrics state from storage
func (metrics *Metrics) Hydrate() error {
	if metrics == nil {
		return fmt.Errorf("cannot hydrate nil reference")
	}
	data, err := metrics.storage.ReadFileFully("metrics.json")
	if err != nil {
		return err
	}
	err = json.Unmarshal(data, metrics)
	if err != nil {
		return err
	}
	return nil
}
