package metrics

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"
	localfs "github.com/jancajthaml-openbank/local-fs"
)

func TestMarshalJSON(t *testing.T) {

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		_, err := json.Marshal(entity)
		if err == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("happy path")
	{
		egress := uint64(10)
		ingress := uint64(20)

		entity := Metrics{
			messageEgress:  &egress,
			messageIngress: &ingress,
		}

		actual, err := json.Marshal(&entity)
		if err != nil {
			t.Fatalf("unexpected error when calling json.Marshal %+v", err)
		}

		aux := &struct {
			MessageEgress  uint64 `json:"messageEgress"`
			MessageIngress uint64 `json:"messageIngress"`
		}{}

		if json.Unmarshal(actual, &aux) != nil {
			t.Errorf("unexpected error when calling json.Unmarshal %+v", err)
		}

		if egress != aux.MessageEgress {
			t.Errorf("extected MessageEgress %d actual %d", egress, aux.MessageEgress)
		}
		if ingress != aux.MessageIngress {
			t.Errorf("extected MessageIngress %d actual %d", egress, aux.MessageIngress)
		}
	}
}

func TestUnmarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		if json.Unmarshal([]byte(""), entity) == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		if json.Unmarshal([]byte(""), entity) == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("error on malformed data")
	{
		egress := uint64(10)
		ingress := uint64(20)

		entity := Metrics{
			messageEgress:  &egress,
			messageIngress: &ingress,
		}

		if json.Unmarshal([]byte("{"), &entity) == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("happy path")
	{
		egress := uint64(10)
		ingress := uint64(20)

		entity := Metrics{
			messageEgress:  &egress,
			messageIngress: &ingress,
		}

		data := []byte("{\"messageEgress\":32,\"messageIngress\":77}")
		err := json.Unmarshal(data, &entity)
		if err != nil {
			t.Fatalf("unexpected error when calling UnmarshalJSON %+v", err)
		}

		if int(*entity.messageEgress) != 32 {
			t.Errorf("extected MessageEgress 32 actual %d", int(*entity.messageEgress))
		}

		if int(*entity.messageIngress) != 77 {
			t.Errorf("extected MessageIngress 77 actual %d", int(*entity.messageIngress))
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

	t.Log("error when marshaling fails")
	{
		entity := Metrics{}
		if entity.Persist() == nil {
			t.Errorf("extected error")
		}
	}

	t.Log("happy path")
	{
		defer os.Remove("/tmp/metrics.json")

		egress := uint64(10)
		ingress := uint64(20)

		entity := Metrics{
			storage:        localfs.NewPlaintextStorage("/tmp"),
			messageEgress:  &egress,
			messageIngress: &ingress,
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

	t.Log("happy path")
	{
		defer os.Remove("/tmp/metrics.json")

		egressOld := uint64(10)
		ingressOld := uint64(20)

		old := Metrics{
			storage:        localfs.NewPlaintextStorage("/tmp"),
			messageEgress:  &egressOld,
			messageIngress: &ingressOld,
		}

		data, err := json.Marshal(&old)
		if err != nil {
			t.Fatalf("unexpected error when calling MarshalJSON %+v", err)
		}

		if ioutil.WriteFile("/tmp/metrics.json", data, 0444) != nil {
			t.Fatalf("unexpected error when writing /tmp/metrics.json")
		}

		egress := uint64(0)
		ingress := uint64(0)

		entity := Metrics{
			storage:        localfs.NewPlaintextStorage("/tmp"),
			messageEgress:  &egress,
			messageIngress: &ingress,
		}

		if entity.Hydrate() != nil {
			t.Fatalf("unexpected error when calling Hydrate")
		}

		if int(*entity.messageEgress) != 10 {
			t.Errorf("extected MessageEgress 10 actual %d", int(*entity.messageEgress))
		}

		if int(*entity.messageIngress) != 20 {
			t.Errorf("extected MessageIngress 20 actual %d", int(*entity.messageIngress))
		}
	}
}
