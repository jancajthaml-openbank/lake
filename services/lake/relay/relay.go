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
	"context"
	"fmt"
	"runtime"
	"time"

	"github.com/jancajthaml-openbank/lake/metrics"
	"github.com/jancajthaml-openbank/lake/support/concurrent"

	zmq "github.com/pebbe/zmq4"
)

// Relay fascade
type Relay struct {
	concurrent.DaemonSupport
	pullPort string
	pubPort  string
	metrics  *metrics.Metrics
}

// NewRelay returns new instance of Relay
func NewRelay(ctx context.Context, pull int, pub int, metrics *metrics.Metrics) *Relay {
	return &Relay{
		DaemonSupport: concurrent.NewDaemonSupport(ctx, "relay"),
		pullPort:      fmt.Sprintf("tcp://127.0.0.1:%d", pull),
		pubPort:       fmt.Sprintf("tcp://127.0.0.1:%d", pub),
		metrics:       metrics,
	}
}

// Start handles everything needed to start relay
func (relay *Relay) Start() {
	if relay == nil {
		return
	}
	var (
		chunk    string
		receiver *zmq.Socket
		sender   *zmq.Socket
		alive    bool = true
	)

	runtime.LockOSThread()
	defer func() {
		recover()
		runtime.UnlockOSThread()
	}()

	ctx, err := zmq.NewContext()
	if err != nil {
		log.Warn().Msgf("Unable to create ZMQ context %+v", err)
		return
	}

	go func() {
		for {
			select {
			case <-relay.Done():
				if !alive {
					return
				}
				alive = false
				relay.MarkDone()
				if ctx.Term() != nil {
					log.Error().Msg("Stop relay-daemon (failed to terminate context)")
				} else {
					log.Info().Msg("Stop relay-daemon")
				}
			}
		}
	}()

	receiver, err = ctx.NewSocket(zmq.PULL)
	if err != nil {
		log.Warn().Msgf("Unable create ZMQ PULL %v", err)
		return
	}

	receiver.SetConflate(false)
	receiver.SetImmediate(true)
	receiver.SetLinger(-1)
	receiver.SetRcvhwm(0)
	defer receiver.Close()

	sender, err = ctx.NewSocket(zmq.PUB)
	if err != nil {
		log.Warn().Msgf("Unable create ZMQ PUB %v", err)
		return
	}

	sender.SetConflate(false)
	sender.SetImmediate(true)
	sender.SetLinger(0)
	sender.SetSndhwm(0)
	defer sender.Close()

	for {
		if receiver.Bind(relay.pullPort) == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	defer receiver.Unbind(relay.pullPort)

	for {
		if sender.Bind(relay.pubPort) == nil {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	defer sender.Unbind(relay.pubPort)

	relay.MarkReady()

	select {
	case <-relay.CanStart:
		break
	case <-relay.Done():
		goto eos
	}

	log.Info().Msg("Start relay-daemon")

loop:
	chunk, err = receiver.Recv(0)
	if err != nil {
		goto fail
	}
	_, err = sender.Send(chunk, 0)
	relay.metrics.MessageIngress()
	if err != nil {
		goto fail
	}
	relay.metrics.MessageEgress()
	goto loop

fail:
	if relay.IsCanceled() {
		goto eos
	}
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
	relay.Stop()
	relay.WaitStop()
	return
}
