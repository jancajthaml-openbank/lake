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

package system

import (
	"fmt"
	"runtime"

	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"

	"github.com/jancajthaml-openbank/lake/pkg/metrics"
	"github.com/jancajthaml-openbank/lake/pkg/utils"
)

// Relay fascade
type Relay struct {
	pullPort      int
	pubPort       int
	metrics       *metrics.Metrics
	alive         bool
	killRequest   chan interface{}
	killConfirmed chan interface{}
}

// New returns new instance of Relay
func New(params utils.RunParams, m *metrics.Metrics) Relay {
	return Relay{
		pullPort:      params.PullPort,
		pubPort:       params.PubPort,
		metrics:       m,
		alive:         false,
		killRequest:   make(chan interface{}),
		killConfirmed: make(chan interface{}),
	}
}

// Stop gracefully autorecovery ZMQ connection
func (r Relay) Stop() {
	r.killRequest <- nil
	<-r.killConfirmed
	log.Info("Stopped ZMQ Relay")

	utils.NotifyServiceStopping()
}

// Start autorecovery ZMQ connection until killed
func (r Relay) Start() {
	r.alive = true

	log.Info("Started ZMQ Relay")
	for {
		work(&r)
		if !r.alive {
			return
		}
		log.Warnf("ZMQ recovering from crash")
	}
}

func isFatalError(err error) bool {
	return err == zmq.ErrorSocketClosed ||
		err == zmq.ErrorContextClosed ||
		zmq.AsErrno(err) == zmq.ETERM
}

func work(r *Relay) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	var (
		chunk       string
		receiver    *zmq.Socket
		sender      *zmq.Socket
		killChannel *zmq.Socket
		pullPort    = fmt.Sprintf("tcp://*:%d", r.pullPort)
		pubPort     = fmt.Sprintf("tcp://*:%d", r.pubPort)
		killPort    = fmt.Sprintf("tcp://127.0.0.1:%d", r.pullPort)
	)

	ctx, err := zmq.NewContext()
	if err != nil {
		log.Warnf("Unable to create ZMQ context %v", err)
		return
	}

	receiver, err = ctx.NewSocket(zmq.PULL)
	if err != nil {
		log.Warnf("Unable create ZMQ PULL %v", err)
		return
	}
	defer receiver.Close()

	sender, err = ctx.NewSocket(zmq.PUB)
	if err != nil {
		log.Warnf("Unable create ZMQ PUB %v", err)
		return
	}
	defer sender.Close()

	killChannel, err = zmq.NewSocket(zmq.PUSH)
	if err != nil {
		log.Warnf("Unable create create ZMQ PUSH %v", err)
		return
	}
	defer killChannel.Close()

	if receiver.Bind(pullPort) != nil {
		log.Warnf("Unable create bind ZMQ PULL")
		return
	}
	defer receiver.Unbind(pullPort)

	if sender.Bind(pubPort) != nil {
		log.Warnf("Unable create bind ZMQ PUB")
		return
	}
	defer sender.Unbind(pubPort)

	if killChannel.Connect(killPort) != nil {
		log.Warnf("Unable to connect kill channel")
		return
	}

	go func() {
		select {
		case <-r.killRequest:
			r.alive = false
			killChannel.Send("", 0)
			return
		}
	}()

	log.Info("ZMQ relay in main loop")

	utils.NotifyServiceReady()

mainLoop:
	chunk, err = receiver.Recv(0)
	switch err {
	case nil:
		if len(chunk) == 0 && !r.alive {
			r.killConfirmed <- nil
			return
		}
		r.metrics.MessageIngress(int64(1))
		sender.Send(chunk, 0)
		r.metrics.MessageEgress(int64(1))
		goto mainLoop
	default:
		if isFatalError(err) {
			log.Warnf("ZMQ crashed in main loop with %v", err)
			return
		}
		goto mainLoop
	}
}
