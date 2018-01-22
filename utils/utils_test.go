package utils

import (
	"context"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"github.com/jancajthaml-openbank/lake/commands"
)

func broadcastClient() *ZMQClient {
	ctx, cancel := context.WithCancel(context.Background())

	client := &ZMQClient{
		push:   make(chan string, 100),
		sub:    make(chan string, 100),
		stop:   cancel,
		host:   "0.0.0.0",
		region: "[",
	}

	go startSubRoutine(ctx, client)
	go startPushRoutine(ctx, client)

	for {
		client.push <- "[]"
		select {
		case <-client.sub:
			log.Info("Broadcast snitcher ready")
			return client
		case <-time.After(10 * time.Millisecond):
			continue
		}
	}
}

func TestZMQClientGracefull(t *testing.T) {
	params := commands.RunParams{
		PullPort: 5562,
		PubPort:  5561,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go commands.RelayMessages(ctx, cancel, params)
	defer func() {
		cancel()
		time.Sleep(100 * time.Millisecond) // INFO give ZMQ time to really unbind
	}()

	t.Log("called public methods on Stopped client")
	{
		client := NewZMQClient("xxx", "0.0.0.0")
		client.Stop()
		client.Publish("yyy", "zzz")
		client.Receive()
	}
}

func TestZMQClientInputValidation(t *testing.T) {
	t.Log("unable to bind to broadcast")
	{
		clientEmpty := NewZMQClient("", "0.0.0.0")
		assert.Nil(t, clientEmpty)

		clientReserved := NewZMQClient("[", "0.0.0.0")
		assert.Nil(t, clientReserved)
	}
}

func TestZMQClientLifecycle(t *testing.T) {
	params := commands.RunParams{
		PullPort: 5562,
		PubPort:  5561,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go commands.RelayMessages(ctx, cancel, params)
	defer func() {
		cancel()
		time.Sleep(100 * time.Millisecond) // INFO give ZMQ time to really unbind
	}()

	client := NewZMQClient("client", "0.0.0.0")

	t.Log("clients communication")
	{
		client.Publish("client", "loopback")
	}

	t.Log("stops properly")
	{
		client.Stop()
		assert.Nil(t, client.Receive())
	}
}

func TestZMQClientCommunication(t *testing.T) {
	params := commands.RunParams{
		PullPort: 5562,
		PubPort:  5561,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go commands.RelayMessages(ctx, cancel, params)
	defer func() {
		cancel()
		time.Sleep(100 * time.Millisecond) // INFO give ZMQ time to really unbind
	}()

	clientFrom := NewZMQClient("from", "0.0.0.0")
	clientTo := NewZMQClient("to", "0.0.0.0")

	t.Log("clients communication")
	{
		clientFrom.Publish("to", "msg-to-from")
		clientTo.Publish("from", "msg-from-to")

		assert.Equal(t, []string{"to", "msg-from-to"}, clientFrom.Receive())
		assert.Equal(t, []string{"from", "msg-to-from"}, clientTo.Receive())
	}

	t.Log("stops properly")
	{
		clientFrom.Stop()
		clientTo.Stop()

		assert.Nil(t, clientTo.Receive())
		assert.Nil(t, clientFrom.Receive())
	}
}

func TestZMQClientBroadcast(t *testing.T) {
	params := commands.RunParams{
		PullPort: 5562,
		PubPort:  5561,
	}

	ctx, cancel := context.WithCancel(context.Background())
	go commands.RelayMessages(ctx, cancel, params)
	defer func() {
		cancel()
		time.Sleep(100 * time.Millisecond) // INFO give ZMQ time to really unbind
	}()

	snitcher := broadcastClient()
	clientA := NewZMQClient("A", "0.0.0.0")
	clientB := NewZMQClient("B", "0.0.0.0")

	t.Log("clients broadcast")
	{
		clientA.Broadcast("public announcement")
		assert.Equal(t, []string{"public", "announcement"}, snitcher.Receive())
	}

	clientA.Stop()
	clientB.Stop()
}
