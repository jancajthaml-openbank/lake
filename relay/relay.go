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

package relay

import (
	"context"
	"fmt"
	"runtime"
	"time"

	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"

	"github.com/jancajthaml-openbank/lake/utils"
)

const backoff = 500 * time.Microsecond

// StartQueue start autorecovery ZMQ connection
func StartQueue(params utils.RunParams) {
	log.Info("Starting ZMQ Relay")

	for {
		ctx, cancel := context.WithCancel(context.Background())
		go RelayMessages(ctx, cancel, params)
		<-ctx.Done()
	}
}

// RelayMessages buffers and relays messages in order
func RelayMessages(ctx context.Context, cancel context.CancelFunc, params utils.RunParams) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer cancel()

	var (
		chunk    string
		receiver *zmq.Socket
		sender   *zmq.Socket
		pullPort = fmt.Sprintf("tcp://*:%d", params.PullPort)
		pubPort  = fmt.Sprintf("tcp://*:%d", params.PubPort)
	)

pullConnection:
	for {
		receiver, err = zmq.NewSocket(zmq.PULL)
		if err == nil {
			break
		}
		if err.Error() == "resource temporarily unavailable" {
			log.Warn("Resources unavailable in connect")
			select {
			case <-time.After(backoff):
				goto pullConnection
			}
		}

		log.Warn("Unable to bind ZMQ socket: ", err)
		return
	}
	defer receiver.Close()

pubConnection:
	for {
		sender, err = zmq.NewSocket(zmq.PUB)
		if err == nil {
			break
		}
		if err.Error() == "resource temporarily unavailable" {
			log.Warn("Resources unavailable in connect")
			select {
			case <-time.After(backoff):
				goto pubConnection
			}
		}

		log.Warn("Unable to bind ZMQ socket: ", err)
		return
	}
	defer sender.Close()

pullBind:
	for {
		err = receiver.Bind(pullPort)
		if err == nil {
			break
		}
		log.Warn("ZMQ receiver unable to bind: ", err)
		select {
		case <-time.After(backoff):
			goto pullBind
		}
	}

pubBind:
	for {
		err = sender.Bind(pubPort)
		if err == nil {
			break
		}
		log.Warn("ZMQ sender unable to bind: ", err)
		select {
		case <-time.After(backoff):
			goto pubBind
		}
	}

	for {
		err = ctx.Err()
		if err != nil {
			return
		}
		chunk, err = receiver.Recv(zmq.DONTWAIT)
		switch err {
		case nil:
			sender.Send(chunk, 0)
			log.Debug(chunk)
		case zmq.ErrorSocketClosed:
			fallthrough
		case zmq.ErrorContextClosed:
			log.Warn("ZMQ connection closed: ", err)
			return
		default:
			continue
		}
	}
}
