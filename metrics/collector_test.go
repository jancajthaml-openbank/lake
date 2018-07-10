package metrics

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMetricsPersist(t *testing.T) {
	entity := NewMetrics()

	t.Log("TimeMessageRelay properly times run of message relay")
	{
		delay := 1e7
		require.Equal(t, int64(0), entity.messageRelayLatency.Count())
		entity.TimeMessageRelay(func() {
			select {
			case <-time.After(time.Duration(delay)):
				return
			}
		})
		assert.Equal(t, int64(1), entity.messageRelayLatency.Count())
		assert.InDelta(t, entity.messageRelayLatency.Percentile(0.95), delay, delay/2)
	}

	t.Log("MessageRelayed properly marks number messages relayed")
	{
		require.Equal(t, int64(0), entity.messagesRelayed.Count())
		entity.MessageRelayed(2)
		assert.Equal(t, int64(2), entity.messagesRelayed.Count())
	}
}
