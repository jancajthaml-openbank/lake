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

package main

import (
	"runtime"

	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
)

// FIXME use tombs
func StartRelay() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	log.Info("Starting ZMQ Relay")

	for {
		relayMessages()
	}
}

func relayMessages() (err error) {
	// FIXME add gracefull crash

	var (
		receiver *zmq.Socket
		sender   *zmq.Socket
	)

pullCreation:
	receiver, err = zmq.NewSocket(zmq.PULL)
	if err != nil && err.Error() == "resource temporarily unavailable" {
		log.Error("Resources unavailable in connect")
		goto pullCreation
	} else if err != nil {
		log.Errorf("Unable to bing ZMQ socket %v", err)
		return
	}
	receiver.SetConflate(true)
	defer receiver.Close()

pubCreation:
	sender, err = zmq.NewSocket(zmq.PUB)
	if err != nil && err.Error() == "resource temporarily unavailable" {
		log.Error("Resources unavailable in connect")
		goto pubCreation
	} else if err != nil {
		log.Errorf("Unable to bing ZMQ socket %v", err)
		return
	}
	sender.SetConflate(false)
	defer sender.Close()

pullConnection:
	if err = receiver.Bind("tcp://*:5562"); err != nil {
		log.Errorf("Unable to bind receiver to ZMQ address %v", err)
		goto pullConnection
	}

pubConnection:
	if err = sender.Bind("tcp://*:5561"); err != nil {
		log.Errorf("Unable to bind sender to ZMQ address %v", err)
		goto pubConnection
	}

sink:
	chunk, err := receiver.RecvBytes(0)
	switch err {
	case nil:
		sender.SendBytes(chunk, 0)
		goto sink
	case zmq.ErrorSocketClosed, zmq.ErrorContextClosed:
		return
	default:
		goto sink
	}
}
