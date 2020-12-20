package relay

import (
	"context"
	"fmt"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	zmq "github.com/pebbe/zmq4"
)


type mockMetrics struct {}
func (_ mockMetrics) MessageEgress() {}
func (_ mockMetrics) MessageIngress() {}


func subRoutine(ctx context.Context, cancel context.CancelFunc, callback chan string, port int) {
	runtime.LockOSThread()
	defer func() {
		cancel()
		runtime.UnlockOSThread()
	}()

	var (
		chunk   string
		channel *zmq.Socket
		err     error
	)

	// FIXME zmq context with proper close and term

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
	defer func() {
		cancel()
		runtime.UnlockOSThread()
	}()

	var (
		channel *zmq.Socket
		err     error
	)

	// FIXME zmq context with proper close and term

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

func TestWorkContract(t *testing.T) {
	metrics := mockMetrics{}

	t.Log("Cancel -> Done")
	{
		relay := NewRelay(5562, 5561, metrics)
		relay.Cancel()
		<-relay.Done()
	}

	t.Log("Setup -> Cancel -> Done")
	{
		relay := NewRelay(5562, 5561, metrics)
		relay.Setup()
		relay.Cancel()
		<-relay.Done()
	}

	t.Log("Setup -> Work -> Cancel -> Done")
	{
		relay := NewRelay(5562, 5561, metrics)
		relay.Setup()
		go relay.Work()
		time.Sleep(100 * time.Millisecond)
		relay.Cancel()
		<-relay.Done()
	}
}

func TestRelayInOrder(t *testing.T) {
	runtime.LockOSThread()
	defer func() {
		recover()
		runtime.UnlockOSThread()
	}()

	metrics := mockMetrics{}
	relay := NewRelay(5562, 5561, metrics)

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

		relay.Setup()
		go relay.Work()

		go pushRoutine(ctx, cancel, pushChannel, 5562)
		go subRoutine(ctx, cancel, subChannel, 5561)

		wg.Add(1)
		go func() {
			defer func() {
				cancel()
				relay.Cancel()
				<-relay.Done()
				wg.Done()
			}()

			for {
				msg := <-subChannel
				accumulatedData = append(accumulatedData, msg)
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
