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

func validRegion(region string) bool {
	if len(region) == 0 || region == "[" {
		log.Warn("invalid region")
		return false
	}
	return true
}

// NewZMQClient returns instance of ZMQClient with both channels ready
func NewZMQClient(region, host string) *ZMQClient {
	if !validRegion(region) {
		return nil
	}

	/*
		conn, err := net.DialTimeout("tcp", host+":"+80, time.Duration(100)*time.Millisecond)
		if err != nil {
			//fmt.Println(err)
			return nil
		}*/

	// FIXME check if host exist or return right there

	ctx, cancel := context.WithCancel(context.Background())
	client := newClient(region, host, cancel)

	go startSubRoutine(ctx, client)
	go startPushRoutine(ctx, client)

	for {
		if client.host == "" {
			return nil
		}
		client.push <- (region + "]")
		select {
		case <-client.sub:
			// FIXME validate that sub is "region + ]"
			log.Infof("ZMQClient(%v) ready", region)
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

	if len(client.host) > 0 {
		client.stop()
		client.push <- ""
		close(client.push)
		close(client.sub)

		client.host = ""

		log.Infof("ZMQClient(%v) closed", client.region)
	}
}

// Broadcast message
func (client *ZMQClient) Broadcast(message string) {
	if client == nil {
		log.Warn("Broadcast called on nil Client")
		return
	}

	if len(client.host) == 0 {
		log.Warnf("ZMQClient(%v) is closed", client.region)
		return
	}

	client.push <- ("[ " + message)
}

// Publish message to remote destination
func (client *ZMQClient) Publish(destination, message string) {
	if len(destination) == 0 {
		client.Broadcast(message)
		return
	}

	if client == nil {
		log.Warn("Publish called on nil Client")
		return
	}

	if len(client.host) == 0 {
		log.Warnf("ZMQClient(%v) is closed", client.region)
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

	if len(client.host) == 0 {
		log.Warnf("ZMQClient(%v) is closed", client.region)
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
