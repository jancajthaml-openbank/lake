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

	"github.com/jancajthaml-openbank/lake/metrics"
	"github.com/jancajthaml-openbank/lake/utils"

	mangos "nanomsg.org/go/mangos/v2"
	"nanomsg.org/go/mangos/v2/protocol/pub"
	"nanomsg.org/go/mangos/v2/protocol/pull"

	_ "nanomsg.org/go/mangos/v2/transport/all"

	log "github.com/sirupsen/logrus"
)

// Relay fascade
type Relay struct {
	utils.DaemonSupport
	pullPort string
	pubPort  string
	metrics  *metrics.Metrics
}

// NewRelay returns new instance of Relay
func NewRelay(ctx context.Context, pull int, pub int, metrics *metrics.Metrics) Relay {
	return Relay{
		DaemonSupport: utils.NewDaemonSupport(ctx, "relay"),
		pullPort:      fmt.Sprintf("tcp://127.0.0.1:%d", pull),
		pubPort:       fmt.Sprintf("tcp://127.0.0.1:%d", pub),
		metrics:       metrics,
	}
}

func worker(relay *Relay, receiver mangos.Socket, sender mangos.Socket) {
	defer relay.Stop()

	var (
		chunk []byte
		err   error
	)

loop:
	chunk, err = receiver.Recv()
	switch err {
	case nil:
		relay.metrics.MessageIngress()
		err = sender.Send(chunk)
		if err != nil {
			log.Warnf("Unable to send message error: %+v", err)
		} else {
			relay.metrics.MessageEgress()
		}
		goto loop
	case mangos.ErrClosed:
		return
	default:
		log.Errorf("%+v", err)
		goto loop
	}
}

// Start handles everything needed to start relay
func (relay Relay) Start() {
	var (
		receiver mangos.Socket
		sender   mangos.Socket
		err      error
		alive    bool = true
	)

	go func() {
		for {
			select {
			case <-relay.Done():
				if !alive {
					return
				}
				alive = false
				if receiver != nil {
					receiver.Close()
				}
				if sender != nil {
					sender.Close()
				}
				log.Info("Stop relay daemon")
				relay.MarkDone()
				return
			}
		}
	}()

	receiver, err = pull.NewSocket()
	if err != nil {
		return
	}

	sender, err = pub.NewSocket()
	if err != nil {
		return
	}

	if receiver.Listen(relay.pullPort) != nil {
		return
	}

	if sender.Listen(relay.pubPort) != nil {
		return
	}

	relay.MarkReady()

	select {
	case <-relay.CanStart:
		break
	case <-relay.Done():
		return
	}

	log.Info("Start relay daemon")

	for id := 0; id < 32; id++ {
		go worker(&relay, receiver, sender)
	}

	<-relay.IsDone
}
