package metrics

import (
	"encoding/json"
	localfs "github.com/jancajthaml-openbank/local-fs"
	"io/ioutil"
	"os"
	"testing"
)

func TestMarshalJSON(t *testing.T) {

	t.Log("does not panic on nil")
	{
		var entity *Metrics
		_, err := entity.MarshalJSON()
		if err == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("happy path")
	{
		entity := Metrics{
			messageEgress:   10,
			messageIngress:  20,
			memoryAllocated: 30,
		}

		actual, err := json.Marshal(&entity)
		if err != nil {
			t.Fatalf("unexpected error when calling json.Marshal %+v", err)
		}

		aux := new(Metrics)

		if json.Unmarshal(actual, aux) != nil {
			t.Errorf("unexpected error when calling json.Unmarshal %+v", err)
		}

		if 10 != aux.messageEgress {
			t.Errorf("extected MessageEgress %d actual %d", 10, aux.messageEgress)
		}
		if 20 != aux.messageIngress {
			t.Errorf("extected MessageIngress %d actual %d", 20, aux.messageIngress)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		if entity.UnmarshalJSON([]byte("")) == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("error on malformed data")
	{
		var entity = new(Metrics)
		if entity.UnmarshalJSON([]byte("{")) == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("happy path")
	{
		entity := Metrics{
			messageEgress:  10,
			messageIngress: 20,
		}

		data := []byte("{\"messageEgress\":32,\"messageIngress\":77}")
		err := json.Unmarshal(data, &entity)
		if err != nil {
			t.Fatalf("unexpected error when calling UnmarshalJSON %+v", err)
		}

		if entity.messageEgress != 32 {
			t.Errorf("extected MessageEgress 32 actual %d", entity.messageEgress)
		}

		if entity.messageIngress != 77 {
			t.Errorf("extected MessageIngress 77 actual %d", entity.messageIngress)
		}
	}
}

func TestPersist(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		if entity.Persist() == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("error when storage is not valid")
	{
		storage, _ := localfs.NewPlaintextStorage("/dev/null")
		entity := Metrics{
			storage:        storage,
			messageEgress:  10,
			messageIngress: 20,
		}
		if entity.Persist() == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("happy path")
	{
		defer os.Remove("/tmp/metrics.json")

		storage, _ := localfs.NewPlaintextStorage("/tmp")
		entity := Metrics{
			storage:        storage,
			messageEgress:  10,
			messageIngress: 20,
		}

		if entity.Persist() != nil {
			t.Fatalf("unexpected error when calling Persist")
		}

		expected, err := json.Marshal(&entity)
		if err != nil {
			t.Fatalf("unexpected error when calling MarshalJSON %+v", err)
		}

		actual, err := ioutil.ReadFile("/tmp/metrics.json")
		if err != nil {
			t.Fatalf("unexpected error when calling reading /tmp/metrics.json")
		}

		if string(expected) != string(actual) {
			t.Errorf("extected %s actual %s", string(expected), string(actual))
		}
	}
}

func TestHydrate(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		if entity.Hydrate() == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("error when set to invalid storage")
	{
		storage, _ := localfs.NewPlaintextStorage("/tmp/a")
		defer os.Remove("/tmp/a")

		entity := Metrics{
			storage: storage,
		}
		if entity.Hydrate() == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("error when file is corrupted")
	{
		defer os.Remove("/tmp/c/metrics.json")

		storage, _ := localfs.NewPlaintextStorage("/tmp/c")
		entity := Metrics{
			storage: storage,
		}
		if ioutil.WriteFile("/tmp/c/metrics.json", []byte("{"), 0644) != nil {
			t.Fatalf("unexpected error when writing /tmp/c/metrics.json")
		}
		if entity.Hydrate() == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("happy path")
	{
		defer os.Remove("/tmp/c/metrics.json")

		storage, _ := localfs.NewPlaintextStorage("/tmp/c")
		old := Metrics{
			storage:        storage,
			messageEgress:  10,
			messageIngress: 20,
		}

		data, err := json.Marshal(&old)
		if err != nil {
			t.Fatalf("unexpected error when calling MarshalJSON %+v", err)
		}

		if ioutil.WriteFile("/tmp/c/metrics.json", data, 0444) != nil {
			t.Fatalf("unexpected error when writing /tmp/c/metrics.json")
		}

		entity := Metrics{
			storage:        storage,
			messageEgress:  0,
			messageIngress: 0,
		}

		if entity.Hydrate() != nil {
			t.Fatalf("unexpected error when calling Hydrate")
		}

		if entity.messageEgress != 10 {
			t.Errorf("extected MessageEgress 10 actual %d", entity.messageEgress)
		}

		if entity.messageIngress != 20 {
			t.Errorf("extected MessageIngress 20 actual %d", entity.messageIngress)
		}
	}
}
