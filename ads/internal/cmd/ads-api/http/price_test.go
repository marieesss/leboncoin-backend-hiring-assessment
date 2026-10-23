package http

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestCentsToEuros(t *testing.T) {
	assert.InDelta(t, 250.0, centsToEuros(25000), 0)
	assert.InDelta(t, 123.45, centsToEuros(12345), 0)
	assert.InDelta(t, 0.99, centsToEuros(99), 0)
	assert.InDelta(t, 0.0, centsToEuros(0), 0)
}

func TestEurosToCents(t *testing.T) {
	t.Run("converts prices", func(t *testing.T) {
		for euros, want := range map[float64]int64{
			250:           25000,
			123.45:        12345,
			19.99:         1999,
			0.1:           10,
			0:             0,
			9_999_999_999: 999_999_999_900,
		} {
			got, err := eurosToCents(euros)

			require.NoError(t, err, euros)
			assert.Equal(t, want, got, euros)
		}
	})

	t.Run("round trips with centsToEuros", func(t *testing.T) {
		for _, cents := range []int64{1, 99, 1999, 34990, 349000} {
			got, err := eurosToCents(centsToEuros(cents))

			require.NoError(t, err, cents)
			assert.Equal(t, cents, got)
		}
	})

	t.Run("error negative price", func(t *testing.T) {
		_, err := eurosToCents(-1)

		require.Error(t, err)
	})

	t.Run("error price too high", func(t *testing.T) {
		_, err := eurosToCents(10_000_000_000)

		require.Error(t, err)
	})
}
