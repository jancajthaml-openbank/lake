package utils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jancajthaml-openbank/lake/commands"
)

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

func TestZMQClientLicecycle(t *testing.T) {
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
	snitch := NewZMQClient("", "0.0.0.0")

	t.Log("clients communication")
	{
		clientFrom.Publish("to", "msg-to-from")
		clientTo.Publish("from", "msg-from-to")

		assert.Equal(t, []string{"to", "msg-from-to"}, clientFrom.Receive())
		assert.Equal(t, []string{"from", "msg-to-from"}, clientTo.Receive())
	}

	t.Log("only expected messages exchanged")
	{
		relayed := [][]string{
			snitch.Receive(),
			snitch.Receive(),
		}

		expected := [][]string{
			{"to", "msg-from-to"},
			{"from", "msg-to-from"},
		}

		assert.ElementsMatch(t, expected, relayed)
	}

	t.Log("stops properly")
	{
		clientFrom.Stop()
		clientTo.Stop()
		snitch.Stop()

		assert.Nil(t, clientTo.Receive())
		assert.Nil(t, clientFrom.Receive())
		assert.Nil(t, snitch.Receive())
	}
}
