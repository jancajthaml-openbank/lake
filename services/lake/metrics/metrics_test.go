package metrics

import (
	"context"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	entity := NewMetrics(ctx, false, "", time.Hour)

	t.Log("MessageEgress properly updates egress messages")
	{
		require.Equal(t, uint64(0), atomic.LoadUint64(entity.messageEgress))

		for i := 1; i <= 10000; i++ {
			entity.MessageEgress()
		}

		assert.Equal(t, uint64(10000), atomic.LoadUint64(entity.messageEgress))
	}

	t.Log("MessageIngress properly updates ingress messages")
	{
		require.Equal(t, uint64(0), atomic.LoadUint64(entity.messageIngress))

		for i := 1; i <= 10000; i++ {
			entity.MessageIngress()
		}

		assert.Equal(t, uint64(10000), atomic.LoadUint64(entity.messageIngress))
	}
}
