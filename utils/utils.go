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

	log "github.com/sirupsen/logrus"
)

const bufferSize = 100

type ZMQClient struct {
	pub     chan string
	sub     chan string
	stop    context.CancelFunc
	running bool
}

func NewZMQClient(channel string, host string) *ZMQClient {
	log.Infof("Creating new client %v", channel)

	ctx, cancel := context.WithCancel(context.Background())

	client := &ZMQClient{
		pub:     make(chan string, bufferSize),
		sub:     make(chan string, bufferSize),
		stop:    cancel,
		running: true,
	}

	go StartZMQPush(ctx, host, channel, client.pub)
	go StartZMQSub(ctx, host, channel, client.sub)

	return client
}

func (client *ZMQClient) Stop() {
	if client == nil {
		log.Warn("Stop called on nil Client")
		return
	}
	if client.running {
		client.stop()
		client.running = false
	}
}

func (client *ZMQClient) Publish(destinationSystem, originSystem, message string) {
	if client == nil {
		log.Warn("Publish called on nil Client")
		return
	}
	client.pub <- (destinationSystem + " " + originSystem + " " + message)
}

func (client *ZMQClient) Receive() []string {
	if client == nil {
		log.Warn("Receive called on nil Client")
		return nil
	}
	return strings.Split(<-client.sub, " ")
}
