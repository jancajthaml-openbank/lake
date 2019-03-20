package daemon

import (
	"fmt"
	"sync"
	"testing"

	"context"
	"runtime"
	"time"

	"github.com/stretchr/testify/assert"

	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"

	"github.com/jancajthaml-openbank/lake/config"
)

func sub(ctx context.Context, cancel context.CancelFunc, callback chan string, port int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer cancel()

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
			log.Warn("Test : Resources unavailable in connect")
			time.Sleep(time.Millisecond)
		} else {
			log.Warn("Test : Unable to connect ZMQ socket", err)
			return
		}
	}
	defer channel.Close()

	for {
		err = channel.Connect(fmt.Sprintf("tcp://0.0.0.0:%d", port))
		if err == nil {
			break
		}
		log.Info("Test : Unable to connect to ZMQ address ", err)
		time.Sleep(time.Millisecond)
	}

	if err = channel.SetSubscribe(""); err != nil {
		log.Warn("Test : Subscription failed ", err)
		return
	}

	for ctx.Err() == nil {
		chunk, err = channel.Recv(0)
		if err != nil {
			if err == zmq.ErrorSocketClosed || err == zmq.ErrorContextClosed {
				log.Info("Test : ZMQ connection closed ", err)
				return
			}
			log.Info("Test : Error while receiving ZMQ message ", err)
			continue
		}
		callback <- chunk
	}
}

func push(ctx context.Context, cancel context.CancelFunc, data chan string, port int) {
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	defer cancel()

	var (
		channel *zmq.Socket
		err     error
	)

	for {
		channel, err = zmq.NewSocket(zmq.PUSH)
		if err == nil {
			break
		}
		if err.Error() == "resource temporarily unavailable" {
			log.Warn("Test : Resources unavailable in connect")
			time.Sleep(time.Millisecond)
		} else {
			log.Warn("Test : Unable to connect ZMQ socket ", err)
			return
		}
	}
	defer channel.Close()

	for {
		err = channel.Connect(fmt.Sprintf("tcp://0.0.0.0:%d", port))
		if err == nil {
			break
		}
		log.Info("Test : Unable to connect to ZMQ address ", err)
		time.Sleep(time.Millisecond)
	}

	for ctx.Err() == nil {
		channel.Send(<-data, 0)
	}
}

func TestRelayInOrder(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	cfg := config.Configuration{
		PullPort: 5562,
		PubPort:  5561,
	}

	metrics := NewMetrics(ctx, cfg)
	relay := NewRelay(ctx, cfg, &metrics)

	t.Log("Relays message")
	{
		accumulatedData := make([]string, 0)
		expectedData := []string{
			"A",
			"B",
			"C",
			"D",
			"E",
			"F",
		}

		pushChannel := make(chan string, len(expectedData))
		subChannel := make(chan string, len(expectedData))

		var wg sync.WaitGroup
		ctx, cancel := context.WithCancel(context.Background())

		go relay.Start()

		select {
		case <-relay.IsReady:
			break
		}

		relay.GreenLight()

		go push(ctx, cancel, pushChannel, cfg.PullPort)
		go sub(ctx, cancel, subChannel, cfg.PubPort)

		wg.Add(1)
		go func() {
			defer func() {
				relay.Stop()
				wg.Done()
			}()

			for {
				accumulatedData = append(accumulatedData, <-subChannel)
				if ctx.Err() == nil && len(expectedData) != len(accumulatedData) {
					continue
				}
				assert.Equal(t, expectedData, accumulatedData)
				return
			}
		}()

		time.Sleep(time.Second)
		for _, msg := range expectedData {
			pushChannel <- msg
		}

		wg.Wait()
	}
}

func TestStartStop(t *testing.T) {
	cfg := config.Configuration{
		PullPort: 5562,
		PubPort:  5561,
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics := NewMetrics(ctx, cfg)
	relay := NewRelay(ctx, cfg, &metrics)

	t.Log("by daemon support ( Start -> Stop )")
	{
		go relay.Start()

		select {
		case <-relay.IsReady:
			break
		}
		relay.GreenLight()

		relay.Stop()
	}
}

func TestStopOnContextCancel(t *testing.T) {
	cfg := config.Configuration{
		PullPort: 5562,
		PubPort:  5561,
	}

	t.Log("stop with cancelation of context")
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		metrics := NewMetrics(ctx, cfg)
		relay := NewRelay(ctx, cfg, &metrics)

		go relay.Start()

		select {
		case <-relay.IsReady:
			break
		}

		relay.GreenLight()

		cancel()
	}
}
