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

package daemon

import (
	"context"
	"fmt"
	"runtime"
	"time"

	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"

	"github.com/jancajthaml-openbank/lake/config"
)

// Relay fascade
type Relay struct {
	Support
	pullPort      string
	pubPort       string
	killPort      string
	metrics       *Metrics
	killConfirmed chan interface{}
	killRequest   chan interface{}
}

// NewRelay returns new instance of Relay
func NewRelay(ctx context.Context, cfg config.Configuration, metrics *Metrics) Relay {
	return Relay{
		Support:       NewDaemonSupport(ctx),
		pullPort:      fmt.Sprintf("tcp://*:%d", cfg.PullPort),
		pubPort:       fmt.Sprintf("tcp://*:%d", cfg.PubPort),
		killPort:      fmt.Sprintf("tcp://127.0.0.1:%d", cfg.PullPort),
		metrics:       metrics,
		killRequest:   make(chan interface{}),
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
	defer relay.MarkDone()
	defer func() { relay.killConfirmed <- nil }()

	alive := true

	go func() {
		select {
		case <-relay.Done():
			if !alive {
				return
			}
			alive = false
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-relay.killConfirmed:
					log.Info("Stop relay daemon")
					return
				case <-ticker.C:
					relay.killRequest <- nil
				}
			}
		}
	}()

	var (
		chunk       string
		receiver    *zmq.Socket
		sender      *zmq.Socket
		killChannel *zmq.Socket
	)

zmqContextNew:
	ctx, err := zmq.NewContext()
	if err != nil {
		log.Warnf("Unable to create ZMQ context %+v", err)
		time.Sleep(10 * time.Millisecond)
		goto zmqContextNew
	}

zmqPullNew:
	receiver, err = ctx.NewSocket(zmq.PULL)
	if err != nil {
		log.Warnf("Unable create ZMQ PULL %v", err)
		time.Sleep(10 * time.Millisecond)
		goto zmqPullNew
	}
	defer receiver.Close()

zmqPubNew:
	sender, err = ctx.NewSocket(zmq.PUB)
	if err != nil {
		log.Warnf("Unable create ZMQ PUB %v", err)
		time.Sleep(10 * time.Millisecond)
		goto zmqPubNew
	}
	defer sender.Close()

zmqKillChannelNew:
	killChannel, err = zmq.NewSocket(zmq.PUSH)
	if err != nil {
		log.Warnf("Unable create create ZMQ PUSH %v", err)
		time.Sleep(10 * time.Millisecond)
		goto zmqKillChannelNew
	}
	defer killChannel.Close()

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

zmqKillChannelConnect:
	if killChannel.Connect(relay.killPort) != nil {
		err = fmt.Errorf("unable to connect kill channel")
		time.Sleep(10 * time.Millisecond)
		goto zmqKillChannelConnect
	}

	relay.MarkReady()

	select {
	case <-relay.canStart:
		break
	case <-relay.Done():
		return
	}

	go func() {
		select {
		case <-relay.killRequest:
			for {
				if killChannel != nil {
					_, err = killChannel.Send("KILL", 0)
					if err == nil {
						return
					}
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

	log.Info("Start relay daemon")

mainLoop:
	chunk, err = receiver.Recv(0)
	switch err {
	case nil:
		if chunk == "KILL" {
			err = nil
			log.Info("Relay killed")
			return
		}
		relay.metrics.MessageIngress(int64(1))

		// FIXME check error
		_, err = sender.Send(chunk, 0)
		if err != nil {
			log.Warnf("Unable to send message error: %+v", err)
		} else {
			relay.metrics.MessageEgress(int64(1))
		}
		goto mainLoop
	default:
		if isFatalError(err) {
			log.Warnf("Relay crashed in main loop with %+v", err)
			return
		}
		goto mainLoop
	}
}

func isFatalError(err error) bool {
	return err == zmq.ErrorSocketClosed ||
		err == zmq.ErrorContextClosed ||
		zmq.AsErrno(err) == zmq.ETERM
}
