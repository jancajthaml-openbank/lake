package relay

import (
	"context"
	"fmt"
	"github.com/jancajthaml-openbank/lake/metrics"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	zmq "github.com/pebbe/zmq4"
)

func subRoutine(ctx context.Context, cancel context.CancelFunc, callback chan string, port int) {
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
			fmt.Println("Test : Resources unavailable in connect")
			time.Sleep(time.Millisecond)
		} else {
			fmt.Printf("Test : Unable to connect SUB socket %+v\n", err)
			return
		}
	}
	defer channel.Close()

	for {
		err = channel.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port))
		if err == nil {
			break
		}
		fmt.Printf("Test : Unable to connect SUB address %+v\n", err)
		time.Sleep(time.Millisecond)
	}

	if err = channel.SetSubscribe(""); err != nil {
		fmt.Printf("Test : Subscription failed %+v\n", err)
		return
	}

	for ctx.Err() == nil {
		chunk, err = channel.Recv(0)
		if err != nil {
			if err == zmq.ErrorSocketClosed || err == zmq.ErrorContextClosed {
				fmt.Printf("Test : ZMQ connection closed %+v\n", err)
				return
			}
			fmt.Printf("Test : Error while receiving ZMQ message %+v\n", err)
			continue
		}
		callback <- chunk
	}
}

func pushRoutine(ctx context.Context, cancel context.CancelFunc, data chan string, port int) {
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
			fmt.Println("Test : Resources unavailable in connect")
			time.Sleep(time.Millisecond)
		} else {
			fmt.Printf("Test : Unable to connect PUSH socket %+v\n", err)
			return
		}
	}
	defer channel.Close()

	for {
		err = channel.Connect(fmt.Sprintf("tcp://127.0.0.1:%d", port))
		if err == nil {
			break
		}
		fmt.Printf("Test : Unable to connect PUSH address %+v\n", err)
		time.Sleep(time.Millisecond)
	}

	for ctx.Err() == nil {
		channel.Send(<-data, 0)
	}
}

func TestStartStop(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	metrics := metrics.NewMetrics(ctx, false, "/tmp", time.Hour)
	relay := NewRelay(ctx, 5562, 5561, metrics)

	t.Log("by daemon support ( Start -> Stop )")
	{
		go relay.Start()
		<-relay.IsReady
		relay.GreenLight()
		relay.Stop()
		relay.WaitStop()
	}
}

func TestStopOnContextCancel(t *testing.T) {
	t.Log("stop with cancelation of context")
	{
		ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)

		metrics := metrics.NewMetrics(ctx, false, "/tmp", time.Hour)
		relay := NewRelay(ctx, 5562, 5561, metrics)

		go relay.Start()
		<-relay.IsReady
		relay.GreenLight()
		cancel()
		relay.WaitStop()
	}
}

func TestRelayInOrder(t *testing.T) {
	masterCtx, masterCancel := context.WithCancel(context.Background())
	defer masterCancel()

	metrics := metrics.NewMetrics(masterCtx, false, "/tmp", time.Hour)
	relay := NewRelay(masterCtx, 5562, 5561, metrics)

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
		<-relay.IsReady
		relay.GreenLight()

		go pushRoutine(ctx, cancel, pushChannel, 5562)
		go subRoutine(ctx, cancel, subChannel, 5561)

		wg.Add(1)
		go func() {
			defer func() {
				relay.Stop()
				relay.WaitStop()
				wg.Done()
			}()

			for {
				accumulatedData = append(accumulatedData, <-subChannel)
				if ctx.Err() == nil && len(expectedData) != len(accumulatedData) {
					continue
				}
				if strings.Join(expectedData, ",") != strings.Join(accumulatedData, ",") {
					t.Errorf("extected %+v actual %+v", expectedData, accumulatedData)
				}
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
