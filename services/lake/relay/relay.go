// Copyright (c) 2016-2019, Jan Cajthaml <jan.cajthaml@gmail.com>
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
	"github.com/jancajthaml-openbank/lake/utils"

	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
)

// Relay fascade
type Relay struct {
	utils.DaemonSupport
	pullPort      string
	pubPort       string
	metrics       *metrics.Metrics
	killConfirmed chan interface{}
}

// NewRelay returns new instance of Relay
func NewRelay(ctx context.Context, pull int, pub int, metrics *metrics.Metrics) Relay {
	return Relay{
		DaemonSupport: utils.NewDaemonSupport(ctx),
		pullPort:      fmt.Sprintf("tcp://*:%d", pull),
		pubPort:       fmt.Sprintf("tcp://*:%d", pub),
		metrics:       metrics,
		killConfirmed: make(chan interface{}),
	}
}

// WaitReady wait for relay to be ready
func (relay Relay) WaitReady(deadline time.Duration) (err error) {
	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("unknown panic")
			}
		}
	}()

	ticker := time.NewTicker(deadline)
	select {
	case <-relay.IsReady:
		ticker.Stop()
		err = nil
		return
	case <-ticker.C:
		err = fmt.Errorf("relay-daemon was not ready within %v seconds", deadline)
		return
	}
}

// Start handles everything needed to start relay
func (relay Relay) Start() {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer recover()
	defer func() { relay.killConfirmed <- nil }()

zmqContextNew:
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Warnf("Unable to create ZMQ context %+v", err)
		time.Sleep(10 * time.Millisecond)
		goto zmqContextNew
	}

	go func() {
		alive := true
		select {
		case <-relay.Done():
			if !alive {
				return
			}
			alive = false
			ctx.Term()
			for {
				select {
				case <-relay.killConfirmed:
					log.Info("Stop relay daemon")
					relay.MarkDone()
					return
				}
			}
		}
	}()

	var (
		chunk    string
		receiver *zmq.Socket
		sender   *zmq.Socket
	)

zmqPullNew:
	receiver, err = ctx.NewSocket(zmq.PULL)
	if err != nil {
		log.Warnf("Unable create ZMQ PULL %v", err)
		time.Sleep(10 * time.Millisecond)
		goto zmqPullNew
	}
	receiver.SetConflate(false)
	receiver.SetImmediate(true)
	receiver.SetLinger(-1)
	receiver.SetRcvhwm(0)
	defer receiver.Close()

zmqPubNew:
	sender, err = ctx.NewSocket(zmq.PUB)
	if err != nil {
		log.Warnf("Unable create ZMQ PUB %v", err)
		time.Sleep(10 * time.Millisecond)
		goto zmqPubNew
	}
	sender.SetConflate(false)
	sender.SetImmediate(true)
	sender.SetLinger(0)
	sender.SetSndhwm(0)
	defer sender.Close()

zmqPullBind:
	if receiver.Bind(relay.pullPort) != nil {
		err = fmt.Errorf("unable create bind ZMQ PULL")
		time.Sleep(10 * time.Millisecond)
		goto zmqPullBind
	}
	defer receiver.Unbind(relay.pullPort)

zmqPubBind:
	if sender.Bind(relay.pubPort) != nil {
		err = fmt.Errorf("unable create bind ZMQ PUB")
		time.Sleep(10 * time.Millisecond)
		goto zmqPubBind
	}
	defer sender.Unbind(relay.pubPort)

	relay.MarkReady()

	select {
	case <-relay.CanStart:
		break
	case <-relay.Done():
		return
	}

	log.Info("Start relay daemon")

mainLoop:
	chunk, err = receiver.Recv(0)
	switch err {
	case nil:
		relay.metrics.MessageIngress()
		_, err = sender.Send(chunk, 0)
		if err != nil {
			if isFatalError(err) {
				log.Warnf("Relay stopping main loop with %+v", err)
				return
			}
			log.Warnf("Unable to send message error: %+v", err)
		} else {
			relay.metrics.MessageEgress()
		}
		goto mainLoop
	default:
		if isFatalError(err) {
			log.Warnf("Relay stopping main loop with %+v", err)
			return
		}
		goto mainLoop
	}
}

func isFatalError(err error) bool {
	return err == zmq.ErrorSocketClosed || err == zmq.ErrorContextClosed ||
		zmq.AsErrno(err) == zmq.ETERM
}
