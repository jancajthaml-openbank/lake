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

	"github.com/jancajthaml-openbank/lake/pkg/metrics"
	"github.com/jancajthaml-openbank/lake/pkg/utils"
)

const ERRRACE = "address already in use "
const ERRBUSSY = "resource temporarily unavailable"
const backoff = 500 * time.Microsecond
const checkRate = 100 * time.Millisecond

// StartQueue start autorecovery ZMQ connection
func StartQueue(params utils.RunParams, m *metrics.Metrics) {
	log.Info("Starting ZMQ Relay")

	for {
		ctx, cancel := context.WithCancel(context.Background())
		go work(ctx, cancel, params, m)
		log.Warn("ZMQ crash, restarting")
		<-ctx.Done()
	}
}

func work(ctx context.Context, cancel context.CancelFunc, params utils.RunParams, m *metrics.Metrics) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var (
		chunk    string
		receiver *zmq.Socket
		sender   *zmq.Socket
		pullPort = fmt.Sprintf("tcp://*:%d", params.PullPort)
		pubPort  = fmt.Sprintf("tcp://*:%d", params.PubPort)
		stopper  = make(chan interface{})
	)

	defer func() {
		cancel()
		<-stopper
		return
	}()

	localCtx, err := zmq.NewContext()
	if err != nil {
		log.Warn("Unable to create ZMQ context:", err)
		return
	}

	ticker := time.NewTicker(checkRate)
	defer ticker.Stop()
	go func() {
		for {
			select {
			case <-ticker.C:
				if ctx.Err() != nil {
					goto teardown
				}
			case <-ctx.Done():
				goto teardown
			}
		}

	teardown:
		localCtx.Term()
		stopper <- nil
		return
	}()

pullConnection:
	receiver, err = localCtx.NewSocket(zmq.PULL)

	switch err {
	case nil:
	case zmq.ErrorSocketClosed, zmq.ErrorContextClosed:
		return
	default:
		log.Warn("Unable create ZMQ PULL connection: ", err)

		if err.Error() == ERRBUSSY {
			log.Warn("Resources unavailable in connect")
			select {
			case <-time.After(backoff):
				goto pullConnection
			}
		}
	}

pubConnection:
	sender, err = localCtx.NewSocket(zmq.PUB)

	switch err {
	case nil:
	case zmq.ErrorSocketClosed, zmq.ErrorContextClosed:
		return
	default:
		log.Warn("Unable create ZMQ PUB connection: ", err)

		if err.Error() == ERRBUSSY {
			log.Warn("Resources unavailable in connect")
			select {
			case <-time.After(backoff):
				goto pubConnection
			}
		}
	}

	if receiver.Bind(pullPort) != nil {
		return
	}

	if sender.Bind(pubPort) != nil {
		return
	}

	log.Info("ZMQ relay in main loop")

mainLoop:
	chunk, err = receiver.Recv(0)
	switch err {
	case nil:
		m.MessageIngress(int64(1))
		sender.Send(chunk, zmq.DONTWAIT)
		m.MessageEgress(int64(1))
		goto mainLoop
	case zmq.ErrorSocketClosed, zmq.ErrorContextClosed:
		return
	}
	goto mainLoop
}
