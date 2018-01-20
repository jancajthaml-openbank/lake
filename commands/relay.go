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

package commands

import (
	"context"
	"runtime"
	"time"

	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
)

const backoff = 50 * time.Millisecond

func StartRelay() {
	log.Info("Starting ZMQ Relay")

	for {
		ctx, cancel := context.WithCancel(context.Background())
		go relayMessages(ctx, cancel)
		<-ctx.Done()
	}
}

func relayMessages(ctx context.Context, cancel context.CancelFunc) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer cancel()

	var (
		chunk    []byte
		receiver *zmq.Socket
		sender   *zmq.Socket
	)

	receiver, err = zmq.NewSocket(zmq.PULL)
	for err != nil {
		if err.Error() == "resource temporarily unavailable" {
			log.Warn("Resources unavailable in connect")
			time.Sleep(backoff)
			receiver, err = zmq.NewSocket(zmq.SUB)
		} else {
			log.Warn("Unable to bind ZMQ socket", err)
			return
		}
	}
	receiver.SetConflate(false)
	defer receiver.Close()

	sender, err = zmq.NewSocket(zmq.PUB)
	for err != nil {
		if err.Error() == "resource temporarily unavailable" {
			log.Warn("Resources unavailable in connect")
			time.Sleep(backoff)
			sender, err = zmq.NewSocket(zmq.SUB)
		} else {
			log.Warn("Unable to bind ZMQ socket", err)
			return
		}
	}
	sender.SetConflate(false)
	defer sender.Close()

	for {
		err = receiver.Bind("tcp://*:5562")
		if err == nil {
			break
		}
		log.Info("Unable to bind reciever to ZMQ address ", err)
		time.Sleep(backoff)
	}

	for {
		err = sender.Bind("tcp://*:5561")
		if err == nil {
			break
		}
		log.Info("Unable to bind sender to ZMQ address", err)
		time.Sleep(backoff)
	}

	for {
		err = ctx.Err()
		if err != nil {
			break
		}
		chunk, err = receiver.RecvBytes(0)
		switch err {
		case nil:
			sender.SendBytes(chunk, 0)
		case zmq.ErrorSocketClosed:
			fallthrough
		case zmq.ErrorContextClosed:
			log.Info("ZMQ connection closed", err)
			return
		default:
			log.Info("Error while receiving ZMQ message", err)
			continue
		}
	}
	return
}
