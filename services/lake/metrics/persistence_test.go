package metrics

import (
	"encoding/json"
	"io/ioutil"
	"os"
	"testing"

	localfs "github.com/jancajthaml-openbank/local-fs"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		_, err := entity.MarshalJSON()
		assert.EqualError(t, err, "cannot marshall nil")
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		_, err := entity.MarshalJSON()
		assert.NotNil(t, err)
	}

	t.Log("happy path")
	{
		egress := uint64(10)
		ingress := uint64(20)

		entity := Metrics{
			messageEgress:  &egress,
			messageIngress: &ingress,
		}

		actual, err := entity.MarshalJSON()
		require.Nil(t, err)

		aux := &struct {
			MessageEgress  uint64 `json:"messageEgress"`
			MessageIngress uint64 `json:"messageIngress"`
		}{}

		require.Nil(t, json.Unmarshal(actual, &aux))

		assert.Equal(t, egress, aux.MessageEgress)
		assert.Equal(t, ingress, aux.MessageIngress)
	}
}

func TestUnmarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		err := entity.UnmarshalJSON([]byte(""))
		assert.NotNil(t, err)
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		err := entity.UnmarshalJSON([]byte(""))
		assert.NotNil(t, err)
	}

	t.Log("error on malformed data")
	{
		egress := uint64(10)
		ingress := uint64(20)

		entity := Metrics{
			messageEgress:  &egress,
			messageIngress: &ingress,
		}

		data := []byte("{")
		assert.NotNil(t, entity.UnmarshalJSON(data))
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
		require.Nil(t, entity.UnmarshalJSON(data))

		assert.NotNil(t, entity.messageEgress)
		assert.Equal(t, 32, int(*entity.messageEgress))

		assert.NotNil(t, entity.messageIngress)
		assert.Equal(t, 77, int(*entity.messageIngress))
	}
}

func TestPersist(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.NotNil(t, entity.Persist())
	}

	t.Log("error when marshaling fails")
	{
		entity := Metrics{}
		assert.NotNil(t, entity.Persist())
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

		require.Nil(t, entity.Persist())

		expected, err := entity.MarshalJSON()
		require.Nil(t, err)

		actual, err := ioutil.ReadFile("/tmp/metrics.json")
		require.Nil(t, err)

		assert.Equal(t, expected, actual)
	}
}

func TestHydrate(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.NotNil(t, entity.Hydrate())
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

		data, err := old.MarshalJSON()
		require.Nil(t, err)

		require.Nil(t, ioutil.WriteFile("/tmp/metrics.json", data, 0444))

		egress := uint64(0)
		ingress := uint64(0)

		entity := Metrics{
			storage:        localfs.NewPlaintextStorage("/tmp"),
			messageEgress:  &egress,
			messageIngress: &ingress,
		}

		require.Nil(t, entity.Hydrate())

		assert.NotNil(t, entity.messageEgress)
		assert.Equal(t, 10, int(*entity.messageEgress))

		assert.NotNil(t, entity.messageIngress)
		assert.Equal(t, 20, int(*entity.messageIngress))
	}
}
