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
	pullPort      int
	pubPort       int
	metrics       Metrics
	killConfirmed chan interface{}
	killRequest   chan interface{}
}

// NewRelay returns new instance of Relay
func NewRelay(ctx context.Context, cfg config.Configuration, metrics Metrics) Relay {
	return Relay{
		Support:       NewDaemonSupport(ctx),
		pullPort:      cfg.PullPort,
		pubPort:       cfg.PubPort,
		metrics:       metrics,
		killRequest:   make(chan interface{}),
		killConfirmed: make(chan interface{}),
	}
}

// Start handles everything needed to start relay
func (r Relay) Start() {
	defer r.MarkDone()
	defer func() { r.killConfirmed <- nil }()

	alive := true

	go func() {
		select {
		case <-r.Done():
			if !alive {
				return
			}
			alive = false
			ticker := time.NewTicker(100 * time.Millisecond)
			defer ticker.Stop()
			for {
				select {
				case <-r.killConfirmed:
					log.Info("Stopped Relay")
					return
				case <-ticker.C:
					r.killRequest <- nil
				}
			}
		}
	}()

	log.Info("Started Relay")

	for {
		err := work(r)
		if !alive {
			return
		}
		if err != nil {
			log.Warnf("Relay recovering from crash %+v", err)
		}
	}
}

func isFatalError(err error) bool {
	return err == zmq.ErrorSocketClosed ||
		err == zmq.ErrorContextClosed ||
		zmq.AsErrno(err) == zmq.ETERM
}

func work(r Relay) (err error) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()

	defer func() {
		if e := recover(); e != nil {
			switch x := e.(type) {
			case string:
				err = fmt.Errorf(x)
			case error:
				err = x
			default:
				err = fmt.Errorf("Unknown panic")
			}
		}
	}()

	var (
		chunk       string
		receiver    *zmq.Socket
		sender      *zmq.Socket
		killChannel *zmq.Socket
		pullPort    = fmt.Sprintf("tcp://*:%d", r.pullPort)
		pubPort     = fmt.Sprintf("tcp://*:%d", r.pubPort)
		killPort    = fmt.Sprintf("tcp://127.0.0.1:%d", r.pullPort)
	)

	go func() {
		select {
		case <-r.killRequest:
			for {
				if killChannel != nil {
					killChannel.Send("KILL", 0)
					return
				}
				time.Sleep(10 * time.Millisecond)
			}
		}
	}()

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
	if receiver.Bind(pullPort) != nil {
		err = fmt.Errorf("Unable create bind ZMQ PULL")
		time.Sleep(10 * time.Millisecond)
		goto zmqPullBind
	}
	defer receiver.Unbind(pullPort)

zmqPubBind:
	if sender.Bind(pubPort) != nil {
		err = fmt.Errorf("Unable create bind ZMQ PUB")
		time.Sleep(10 * time.Millisecond)
		goto zmqPubBind
	}
	defer sender.Unbind(pubPort)

zmqKillChannelConnect:
	if killChannel.Connect(killPort) != nil {
		err = fmt.Errorf("Unable to connect kill channel")
		time.Sleep(10 * time.Millisecond)
		goto zmqKillChannelConnect
	}

	log.Info("Relay in main loop")

	r.MarkReady()

mainLoop:
	chunk, err = receiver.Recv(0)
	switch err {
	case nil:
		if chunk == "KILL" {
			err = nil
			log.Info("Relay killed")
			return
		}
		r.metrics.MessageIngress(int64(1))
		sender.Send(chunk, 0)
		r.metrics.MessageEgress(int64(1))
		goto mainLoop
	default:
		if isFatalError(err) {
			log.Warnf("Relay crashed in main loop with %+v", err)
			return
		}
		goto mainLoop
	}
}
