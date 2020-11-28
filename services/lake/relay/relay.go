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

package relay

import (
	"fmt"
	"time"
	"github.com/jancajthaml-openbank/lake/metrics"
	"github.com/jancajthaml-openbank/lake/support/concurrent"

	zmq "github.com/pebbe/zmq4"
)

// Relay fascade
type Relay struct {
	pullPort string
	pubPort  string
	metrics  *metrics.Metrics
	receiver *zmq.Socket
	sender   *zmq.Socket
	ctx      *zmq.Context
	done     chan(interface{})
}

// NewRelay returns new instance of Relay
func NewRelay(pull int, pub int, metrics *metrics.Metrics) concurrent.Worker {
	return &Relay{
		pullPort:      fmt.Sprintf("tcp://127.0.0.1:%d", pull),
		pubPort:       fmt.Sprintf("tcp://127.0.0.1:%d", pub),
		metrics:       metrics,
		done:          make(chan interface{}),
	}
}

func (relay *Relay) Setup() error {
	if relay == nil {
		return fmt.Errorf("nil pointer")
	}

	var err error
	relay.ctx, err = zmq.NewContext()
	if err != nil {
		return fmt.Errorf("unable to create ZMQ context %+v", err)
	}

	relay.receiver, err = relay.ctx.NewSocket(zmq.PULL)
	if err != nil {
		return fmt.Errorf("unable create ZMQ PULL %v", err)
	}

	relay.receiver.SetConflate(false)
	relay.receiver.SetImmediate(true)
	relay.receiver.SetLinger(-1)
	relay.receiver.SetRcvhwm(0)

	relay.sender, err = relay.ctx.NewSocket(zmq.PUB)
	if err != nil {
		return fmt.Errorf("unable create ZMQ PUB %v", err)
	}

	relay.sender.SetConflate(false)
	relay.sender.SetImmediate(true)
	relay.sender.SetLinger(0)
	relay.sender.SetSndhwm(0)

	for {
		if relay.receiver.Bind(relay.pullPort) == nil {
			break
		}
		relay.receiver.Unbind(relay.pullPort)
		time.Sleep(10 * time.Millisecond)
	}

	for {
		if relay.sender.Bind(relay.pubPort) == nil {
			break
		}
		relay.sender.Unbind(relay.pubPort)
		time.Sleep(10 * time.Millisecond)
	}
	return nil
}

func (relay *Relay) Cancel() {
	if relay.sender != nil {
		relay.sender.Unbind(relay.pubPort)
		relay.sender.Close()
	}
	if relay.receiver != nil {
		relay.receiver.Unbind(relay.pullPort)
		relay.receiver.Close()
	}
	if relay.ctx != nil {
		for relay.ctx.Term() != nil {}
	}
	relay.sender = nil
	relay.receiver = nil
	relay.ctx = nil
}

func (relay *Relay) Done() <- chan interface{} {
	return relay.done
}

// Start handles everything needed to start relay
func (relay *Relay) Work() {
	if relay == nil {
		return
	}

	defer func() {
		recover()
		close(relay.done)
	}()

	var chunk string
	var err error

loop:
	chunk, err = relay.receiver.Recv(0)
	if err != nil {
		goto fail
	}
	_, err = relay.sender.Send(chunk, 0)
	relay.metrics.MessageIngress()
	if err != nil {
		goto fail
	}
	relay.metrics.MessageEgress()
	goto loop

fail:
	if err == nil {
		goto loop
	}
	if err == zmq.ErrorSocketClosed || err == zmq.ErrorContextClosed {
		goto eos
	}
	if zmq.AsErrno(err) == zmq.ETERM {
		goto eos
	}
	goto loop

eos:
	return
}
