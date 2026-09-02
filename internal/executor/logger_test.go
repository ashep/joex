package executor

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_Write(main *testing.T) {
	main.Parallel()

	main.Run("NoScroll", func(t *testing.T) {
		t.Parallel()

		l := newLogger(6)

		n, err := l.Write([]byte("123"))
		require.NoError(t, err)
		assert.Equal(t, 3, n)
		assert.Equal(t, "123", l.String())

		n, err = l.Write([]byte("abc"))
		require.NoError(t, err)
		assert.Equal(t, 3, n)
		assert.Equal(t, "123abc", l.String())
	})

	main.Run("Scroll", func(t *testing.T) {
		t.Parallel()

		l := newLogger(5)
		_, _ = l.Write([]byte("123"))
		n, err := l.Write([]byte("abc"))

		require.NoError(t, err)
		assert.Equal(t, 3, n)
		assert.Equal(t, "23abc", l.String())
	})

	main.Run("OversizedMessage", func(t *testing.T) {
		t.Parallel()

		l := newLogger(3)
		n, err := l.Write([]byte("12345"))

		require.NoError(t, err)
		assert.Equal(t, 3, n)
		assert.Equal(t, "345", l.String())
	})
}
