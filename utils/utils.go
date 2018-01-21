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

package utils

import (
	"context"
	"strings"
	"time"

	log "github.com/sirupsen/logrus"
)

const bufferSize = 100

// ZMQClient is a fascade for ZMQ queue
type ZMQClient struct {
	push    chan string
	sub     chan string
	stop    context.CancelFunc
	running bool
	region  string
}

// NewZMQClient returns instance of ZMQClient with both channels ready
func NewZMQClient(region, host string) *ZMQClient {
	ctx, cancel := context.WithCancel(context.Background())

	client := &ZMQClient{
		push:    make(chan string, bufferSize),
		sub:     make(chan string, bufferSize),
		stop:    cancel,
		running: true,
		region:  region,
	}

	go startSubRoutine(ctx, host, region, client.sub)
	go startPushRoutine(ctx, host, region, client.push)

	for {
		client.push <- (region + "]")
		select {
		case <-client.sub:
			log.Infof("ZMQ Client \"%v\" ready", region)
			return client
		case <-time.After(10 * time.Millisecond):
			continue
		}
	}
}

// Stop ZMQ connections and close ZMQClient channels
func (client *ZMQClient) Stop() {
	if client == nil {
		log.Warn("Stop called on nil Client")
		return
	}

	if !client.running {
		log.Warnf("ZMQ Client \"%v\" is closed", client.region)
		return
	}

	if client.running {
		client.stop()
		close(client.push)
		close(client.sub)

		client.running = false

		log.Infof("ZMQ Client \"%v\" closed", client.region)
	}
}

// Publish message to remote destination
func (client *ZMQClient) Publish(destination, message string) {
	if client == nil {
		log.Warn("Publish called on nil Client")
		return
	}

	if !client.running {
		log.Warnf("ZMQ Client \"%v\" is closed", client.region)
		return
	}

	client.push <- (destination + " " + client.region + " " + message)
}

// Receive message for this region
func (client *ZMQClient) Receive() []string {
	if client == nil {
		log.Warn("Receive called on nil Client")
		return nil
	}

	if !client.running {
		log.Warnf("ZMQ Client \"%v\" is closed", client.region)
		return nil
	}

	for {
		data := <-client.sub
		if data == (client.region + "]") {
			continue
		}
		return strings.Split(data, " ")[1:]
	}
}
