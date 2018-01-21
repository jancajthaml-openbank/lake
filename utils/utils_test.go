package utils

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"github.com/jancajthaml-openbank/lake/commands"
)

func TestZMQClientGracefull(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	go commands.RelayMessages(ctx, cancel)
	defer cancel()

	t.Log("called publis methods on Stopped client")
	{
		client := NewZMQClient("xxx", "0.0.0.0")
		client.Stop()
		client.Publish("xxx", "yyy", "zzz")
		client.Receive()
	}
}

func TestZMQClientLicecycle(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	go commands.RelayMessages(ctx, cancel)
	defer cancel()

	clientFrom := NewZMQClient("from", "0.0.0.0")
	clientTo := NewZMQClient("to", "0.0.0.0")
	snitch := NewZMQClient("", "0.0.0.0")

	t.Log("verify topics isolation")
	{
		clientFrom.Publish("to", "from", "msg-to-from")
		clientTo.Publish("from", "to", "msg-from-to")

		assert.Equal(t, []string{"from", "to", "msg-from-to"}, clientFrom.Receive())
		assert.Equal(t, []string{"to", "from", "msg-to-from"}, clientTo.Receive())
	}

	t.Log("verify messages format")
	{
		relayed := [][]string{
			snitch.Receive(),
			snitch.Receive(),
		}

		expected := [][]string{
			{"from", "to", "msg-from-to"},
			{"to", "from", "msg-to-from"},
		}

		assert.ElementsMatch(t, expected, relayed)
	}

	clientFrom.Stop()
	clientTo.Stop()
	snitch.Stop()

	assert.Nil(t, clientTo.Receive())
	assert.Nil(t, clientFrom.Receive())
	assert.Nil(t, snitch.Receive())
}
