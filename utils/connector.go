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

package utils

import (
	"context"
	"fmt"
	"runtime"
	"time"

	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
)

const backoff = 50 * time.Millisecond

func startSubRoutine(master context.Context, client *ZMQClient) {
	log.Debugf("ZMQ SUB %s work", client.region)

	for {
		ctx, cancel := context.WithCancel(master)
		go workZMQSub(ctx, cancel, client)
		<-ctx.Done()
		if master.Err() != nil {
			break
		}
	}
}

func startPushRoutine(master context.Context, client *ZMQClient) {
	log.Debugf("ZMQ PUSH %s work", client.region)

	for {
		ctx, cancel := context.WithCancel(master)
		go workZMQPush(ctx, cancel, client)
		<-ctx.Done()
		if master.Err() != nil {
			break
		}
	}
}

func workZMQSub(ctx context.Context, cancel context.CancelFunc, client *ZMQClient) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer cancel()
	defer func() {
		recover()
	}()

	var (
		chunk   string
		channel *zmq.Socket
		err     error
	)

	for {
		channel, err = zmq.NewSocket(zmq.SUB)
		if err == nil {
			break
		}
		if err.Error() == "resource temporarily unavailable" {
			log.Warn("Resources unavailable in connect")
			time.Sleep(backoff)
		} else {
			log.Warn("Unable to connect ZMQ socket ", err)
			return
		}
	}
	channel.SetConflate(false)
	defer channel.Close()

	for {
		err = channel.Connect(fmt.Sprintf("tcp://%s:%d", client.host, 5561))
		if err == nil {
			// FIXME check error here
			break
		}
		log.Info("Unable to connect to ZMQ address ", err)
		time.Sleep(backoff)
	}

	if err = channel.SetSubscribe(client.region); err != nil {
		log.Warn("Subscription to %s failed ", client.region, err)
		return
	}

	for {
		err = ctx.Err()
		if err != nil {
			break
		}

		chunk, err = channel.Recv(0)
		switch err {
		case nil:
			client.sub <- chunk
		case zmq.ErrorSocketClosed:
			fallthrough
		case zmq.ErrorContextClosed:
			log.Info("ZMQ connection closed ", err)
			return
		default:
			log.Info("Error while receiving ZMQ message ", err)
			continue
		}
	}
}

func workZMQPush(ctx context.Context, cancel context.CancelFunc, client *ZMQClient) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer cancel()
	defer func() {
		recover()
	}()

	var (
		chunk   string
		channel *zmq.Socket
		err     error
	)

	for {
		channel, err = zmq.NewSocket(zmq.PUSH)
		if err == nil {
			break
		}
		if err.Error() == "resource temporarily unavailable" {
			log.Warn("Resources unavailable in connect")
			time.Sleep(backoff)
		} else {
			log.Warn("Unable to connect ZMQ socket ", err)
			return
		}
	}
	channel.SetConflate(false)
	defer channel.Close()

	for {
		err = channel.Connect(fmt.Sprintf("tcp://%s:%d", client.host, 5562))
		if err == nil {
			break
		}
		log.Info("Unable to connect to ZMQ address ", err)
		time.Sleep(backoff)
	}

	for {
		chunk = <-client.push
		err = ctx.Err()
		if err != nil {
			break
		}
		channel.Send(chunk, 0)
	}
}
