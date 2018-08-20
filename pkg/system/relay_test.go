package system

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

	"github.com/jancajthaml-openbank/lake/pkg/metrics"
	"github.com/jancajthaml-openbank/lake/pkg/utils"
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
	params := utils.RunParams{
		PullPort: 5562,
		PubPort:  5561,
	}

	m := metrics.NewMetrics()

	relay := NewRelay(params, m)

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
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		go relay.Start()
		go push(ctx, cancel, pushChannel, params.PullPort)
		go sub(ctx, cancel, subChannel, params.PubPort)

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer relay.Stop()
			defer cancel()

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
	params := utils.RunParams{
		PullPort: 5562,
		PubPort:  5561,
	}

	m := metrics.NewMetrics()

	relay := NewRelay(params, m)

	t.Log("by API ( Start->Stop )")
	{
		go relay.Start()
		time.Sleep(300 * time.Millisecond)
		relay.Stop()
	}
}

func TestAbleToRelayMessagesAfterCrash(t *testing.T) {
	params := utils.RunParams{
		PullPort: 5562,
		PubPort:  5561,
	}

	m := metrics.NewMetrics()

	relay := NewRelay(params, m)

	t.Log("ZMQ is able to relay messages after it was stopped")
	{
		go relay.Start()
		time.Sleep(300 * time.Millisecond)
		relay.Stop()

		accumulatedData := make([]string, 0)
		expectedData := []string{"X", "Y"}

		pushChannel := make(chan string, len(expectedData))
		subChannel := make(chan string, len(expectedData))

		var wg sync.WaitGroup
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		go relay.Start()
		go push(ctx, cancel, pushChannel, params.PullPort)
		go sub(ctx, cancel, subChannel, params.PubPort)

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer relay.Stop()
			defer cancel()

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
