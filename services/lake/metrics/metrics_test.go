package metrics

import (
	"context"
	"testing"
	"time"
)

func TestMetricsIngress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()


	t.Log("noop on nil reference")
	{
		var entity *Metrics
		entity.MessageIngress()
	}

	t.Log("properly updates ingress messages")
	{
		entity := NewMetrics(ctx, false, "/tmp/ingress", time.Hour)

		if uint64(0) != entity.messageIngress {
			t.Errorf("extected messageIngress 0 actual %d", entity.messageIngress)
		}

		for i := 1; i <= 1000; i++ {
			entity.MessageIngress()
		}

		if uint64(1000) != entity.messageIngress {
			t.Errorf("extected messageIngress 10000 actual %d", entity.messageIngress)
		}
	}
}

func TestMetricsEgress(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Log("noop on nil reference")
	{
		var entity *Metrics
		entity.MessageEgress()
	}

	t.Log("properly updates egress messages")
	{
		entity := NewMetrics(ctx, false, "/tmp/egress", time.Hour)

		if uint64(0) != entity.messageEgress {
			t.Errorf("extected messageEgress 0 actual %d", entity.messageEgress)
		}

		for i := 1; i <= 1000; i++ {
			entity.MessageEgress()
		}

		if uint64(1000) != entity.messageEgress {
			t.Errorf("extected messageEgress 10000 actual %d", entity.messageEgress)
		}
	}
}

func TestMetricsMemory(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	t.Log("noop on nil reference")
	{
		var entity *Metrics
		entity.MemoryAllocatedSnapshot()
	}

	t.Log("properly updates ingress messages")
	{
		entity := NewMetrics(ctx, false, "/tmp/memory", time.Hour)

		if uint64(0) != entity.memoryAllocated {
			t.Errorf("extected memoryAllocated 0 actual %d", entity.memoryAllocated)
		}

		entity.MemoryAllocatedSnapshot()

		if uint64(0) == entity.memoryAllocated {
			t.Errorf("extected memoryAllocated to be non zero after snapshot")
		}
	}
}
