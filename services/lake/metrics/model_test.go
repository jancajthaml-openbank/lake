package metrics

import (
	"testing"

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
		assert.EqualError(t, err, "cannot marshall nil references")
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

		data := []byte("{\"messageEgress\":10,\"messageIngress\":20}")

		assert.Equal(t, data, actual)
	}
}

func TestUnmarshalJSON(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		err := entity.UnmarshalJSON([]byte(""))
		assert.EqualError(t, err, "cannot unmarshall to nil")
	}

	t.Log("error when values are nil")
	{
		entity := Metrics{}
		err := entity.UnmarshalJSON([]byte(""))
		assert.EqualError(t, err, "cannot unmarshall to nil references")
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
