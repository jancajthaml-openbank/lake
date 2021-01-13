// Copyright (c) 2016-2021, Jan Cajthaml <jan.cajthaml@gmail.com>
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

	"github.com/pebbe/zmq4"
)

// Relay 1:N (PULL -> PUB)
type Relay struct {
	pullPort  string
	pubPort   string
	metrics   metrics.Metrics
	puller    *zmq4.Socket
	pusher    *zmq4.Socket
	publisher *zmq4.Socket
	ctx       *zmq4.Context
	done      chan interface{}
	live      bool
}

// NewRelay returns new instance of Relay
func NewRelay(pull int, pub int, metrics metrics.Metrics) *Relay {
	return &Relay{
		pullPort: fmt.Sprintf("tcp://127.0.0.1:%d", pull),
		pubPort:  fmt.Sprintf("tcp://127.0.0.1:%d", pub),
		metrics:  metrics,
		done:     nil,
		live:     false,
	}
}

func (relay *Relay) setupContext() (err error) {
	if relay == nil || relay.ctx != nil {
		return
	}
	relay.ctx, err = zmq4.NewContext()
	if err != nil {
		return
	}
	relay.ctx.SetRetryAfterEINTR(false)
	return
}

func (relay *Relay) setupPuller() (err error) {
	if relay == nil || relay.puller != nil {
		return
	}
	relay.puller, err = relay.ctx.NewSocket(zmq4.PULL)
	if err != nil {
		return
	}
	relay.puller.SetConflate(false)
	relay.puller.SetImmediate(true)
	relay.puller.SetLinger(0)
	relay.puller.SetRcvhwm(0)
	for relay.puller.Bind(relay.pullPort) != nil {
	}
	return
}

func (relay *Relay) setupPublisher() (err error) {
	if relay == nil || relay.publisher != nil {
		return
	}
	relay.publisher, err = relay.ctx.NewSocket(zmq4.PUB)
	if err != nil {
		return
	}
	relay.publisher.SetConflate(false)
	relay.publisher.SetImmediate(true)
	relay.publisher.SetLinger(0)
	relay.publisher.SetSndhwm(0)
	relay.publisher.SetXpubNodrop(true)
	for relay.publisher.Bind(relay.pubPort) != nil {
	}
	return
}

func (relay *Relay) setupPusher() (err error) {
	if relay == nil || relay.pusher != nil {
		return
	}
	relay.pusher, err = relay.ctx.NewSocket(zmq4.PUSH)
	if err != nil {
		return
	}
	for relay.pusher.Connect(relay.pullPort) != nil {
	}
	return
}

// Setup initializes zmq context and sockets
func (relay *Relay) Setup() error {
	if relay == nil {
		return fmt.Errorf("nil pointer")
	}
	var err error
	err = relay.setupContext()
	if err != nil {
		return fmt.Errorf("unable to create context %+v", err)
	}
	err = relay.setupPuller()
	if err != nil {
		return fmt.Errorf("unable create PULL socket %v", err)
	}
	err = relay.setupPublisher()
	if err != nil {
		return fmt.Errorf("unable create PUB socket %v", err)
	}
	err = relay.setupPusher()
	if err != nil {
		return fmt.Errorf("unable create PUSH socket %v", err)
	}
	return nil
}

// Cancel shut downs sockets and terminates context
func (relay *Relay) Cancel() {
	if relay == nil {
		return
	}
	if relay.pusher != nil && relay.live {
		relay.live = false
		relay.pusher.SendBytes([]byte("_"), 0)
	}
	<-relay.Done()
	if relay.publisher != nil {
		relay.publisher.Close()
	}
	if relay.puller != nil {
		relay.puller.Close()
	}
	if relay.pusher != nil {
		relay.pusher.Close()
	}
	relay.publisher = nil
	relay.puller = nil
	relay.pusher = nil
	if relay.ctx != nil {
		relay.ctx.Term()
	}
	relay.ctx = nil
}

// Done returns done when relay is finished if nil returns done immediately
func (relay *Relay) Done() <-chan interface{} {
	if relay == nil || relay.done == nil {
		done := make(chan interface{})
		close(done)
		return done
	}
	return relay.done
}

// Work runs relay main loop
func (relay *Relay) Work() {
	if relay == nil {
		return
	}

	relay.live = true
	relay.done = make(chan interface{})

	defer func() {
		recover()
		close(relay.done)
	}()

	var chunk []byte
	var err error

pull:
	chunk, err = relay.puller.RecvBytes(0)
	if err != nil {
		log.Warn().Msgf("Unable to receive message with %+v", err)
		goto fail
	}
	if !relay.live {
		goto eos
	}
pub:
	relay.metrics.MessageIngress()
	_, err = relay.publisher.SendBytes(chunk, 0)
	if err != nil {
		log.Warn().Msgf("Unable to send message with %+v", err)
		if err.Error() != "resource temporarily unavailable" {
			time.Sleep(10*time.Millisecond)
			goto pub
		}
		goto fail
	}
	relay.metrics.MessageEgress()
	_, err = relay.publisher.SendBytes(chunk, 0)
	if err != nil {
		log.Warn().Msgf("Unable to send message with %+v", err)
		if err.Error() != "resource temporarily unavailable" {
			goto pub
		}
		goto fail
	}
	relay.metrics.MessageEgress()
	log.Debug().Msgf("Relayed %s", string(chunk))
	goto pull

fail:
	switch err {
	case zmq4.ErrorNoSocket:
		fallthrough
	case zmq4.ErrorSocketClosed:
		fallthrough
	case zmq4.ErrorContextClosed:
		goto eos
	default:
		goto pull
	}

eos:
	return
}
