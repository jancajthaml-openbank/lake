package commands

import (
	"sync"
	"testing"

	"context"
	"runtime"
	"time"

	"github.com/stretchr/testify/assert"

	zmq "github.com/pebbe/zmq4"
	log "github.com/sirupsen/logrus"
)

func sub(ctx context.Context, cancel context.CancelFunc, callback chan string) {
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
		err = channel.Connect("tcp://0.0.0.0:5561")
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

func push(ctx context.Context, cancel context.CancelFunc, data chan string) {
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
		err = channel.Connect("tcp://0.0.0.0:5562")
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

		go RelayMessages(ctx, cancel)
		go push(ctx, cancel, pushChannel)
		go sub(ctx, cancel, subChannel)

		wg.Add(1)
		go func() {
			defer wg.Done()
			defer cancel()

			for {
				accumulatedData = append(accumulatedData, <-subChannel)
				time.Sleep(time.Millisecond)
				if ctx.Err() == nil && len(expectedData) != len(accumulatedData) {
					continue
				}
				assert.Equal(t, expectedData, accumulatedData)
				return
			}
		}()

		time.Sleep(300 * time.Millisecond)
		for _, msg := range expectedData {
			pushChannel <- msg
		}

		wg.Wait()
	}
}

/*
func BenchmarkRelay(b *testing.B) {
	capacity := 1000
	pushChannel := make(chan string, capacity)
	subChannel := make(chan string, capacity)
	msg := "aaaaaaaa"

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go push(ctx, cancel, pushChannel)
	go sub(ctx, cancel, subChannel)

	b.ResetTimer()
	b.SetBytes(376)

	for i := 0; i < b.N; i++ {
		pushChannel <- msg
		<-subChannel
	}
}*/
