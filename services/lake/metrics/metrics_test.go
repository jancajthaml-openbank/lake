package metrics

import "testing"

func TestMetrics(t *testing.T) {
	t.Log("does not panic on invalid root")
	{
		NewMetrics("/dev/null", false)
	}

	t.Log("does not panic on nil reference function calls")
	{
		var entity *Metrics
		entity.MessageIngress()
		entity.MessageEgress()
		entity.MemoryAllocatedSnapshot()
	}
}

func TestMetricsIngress(t *testing.T) {
	entity := NewMetrics("/tmp/ingress", false)

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

func TestMetricsEgress(t *testing.T) {
	entity := NewMetrics("/tmp/egress", false)

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

func TestMetricsMemory(t *testing.T) {
	entity := NewMetrics("/tmp/memory", false)

	if uint64(0) != entity.memoryAllocated {
		t.Errorf("extected memoryAllocated 0 actual %d", entity.memoryAllocated)
	}

	entity.MemoryAllocatedSnapshot()

	if uint64(0) == entity.memoryAllocated {
		t.Errorf("extected memoryAllocated to be non zero after snapshot")
	}
}
