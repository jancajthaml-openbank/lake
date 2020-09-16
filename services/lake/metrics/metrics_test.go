package metrics

import (
	"context"
	"sync/atomic"
	"testing"
	"time"
)

func TestMetrics(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	entity := NewMetrics(ctx, false, "/tmp", time.Hour)

	t.Log("MessageEgress properly updates egress messages")
	{
		if uint64(0) != atomic.LoadUint64(entity.messageEgress) {
			t.Errorf("extected MessageEgress 0 actual %d", atomic.LoadUint64(entity.messageEgress))
		}

		for i := 1; i <= 10000; i++ {
			entity.MessageEgress()
		}

		if uint64(10000) != atomic.LoadUint64(entity.messageEgress) {
			t.Errorf("extected MessageEgress 10000 actual %d", atomic.LoadUint64(entity.messageEgress))
		}
	}

	t.Log("MessageIngress properly updates ingress messages")
	{
		if uint64(0) != atomic.LoadUint64(entity.messageIngress) {
			t.Errorf("extected MessageIngress 0 actual %d", atomic.LoadUint64(entity.messageIngress))
		}

		for i := 1; i <= 10000; i++ {
			entity.MessageIngress()
		}

		if uint64(10000) != atomic.LoadUint64(entity.messageIngress) {
			t.Errorf("extected MessageIngress 10000 actual %d", atomic.LoadUint64(entity.messageIngress))
		}
	}
}
