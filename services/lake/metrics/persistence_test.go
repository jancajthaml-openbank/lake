package metrics

import (
	"io/ioutil"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPersist(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.EqualError(t, entity.Persist(), "cannot persist nil reference")
	}

	t.Log("error when marshalling fails")
	{
		entity := Metrics{}
		assert.EqualError(t, entity.Persist(), "json: error calling MarshalJSON for type *metrics.Metrics: cannot marshall nil references")
	}

	t.Log("error when race cannot open tempfile for writing")
	{
		egress := uint64(10)
		ingress := uint64(20)

		entity := Metrics{
			output:         "/sys/kernel/security",
			messageEgress:  &egress,
			messageIngress: &ingress,
		}

		assert.NotNil(t, entity.Persist())
	}

	t.Log("happy path")
	{
		tmpfile, err := ioutil.TempFile(os.TempDir(), "test_metrics_persist")

		require.Nil(t, err)
		defer os.Remove(tmpfile.Name())

		egress := uint64(10)
		ingress := uint64(20)

		entity := Metrics{
			output:         tmpfile.Name(),
			messageEgress:  &egress,
			messageIngress: &ingress,
		}

		require.Nil(t, entity.Persist())

		expected, err := entity.MarshalJSON()
		require.Nil(t, err)

		actual, err := ioutil.ReadFile(tmpfile.Name())
		require.Nil(t, err)

		assert.Equal(t, expected, actual)
	}
}

func TestHydrate(t *testing.T) {

	t.Log("error when caller is nil")
	{
		var entity *Metrics
		assert.EqualError(t, entity.Hydrate(), "cannot hydrate nil reference")
	}

	t.Log("happy path")
	{
		tmpfile, err := ioutil.TempFile(os.TempDir(), "test_metrics_hydrate")

		require.Nil(t, err)
		defer os.Remove(tmpfile.Name())

		egress_old := uint64(10)
		ingress_old := uint64(20)

		old := Metrics{
			messageEgress:  &egress_old,
			messageIngress: &ingress_old,
		}

		data, err := old.MarshalJSON()
		require.Nil(t, err)

		require.Nil(t, ioutil.WriteFile(tmpfile.Name(), data, 0444))

		egress := uint64(0)
		ingress := uint64(0)

		entity := Metrics{
			output:         tmpfile.Name(),
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
