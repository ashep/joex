package datatype_test

import (
	"testing"

	"github.com/ashep/joex/internal/datatype"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTypeError_Error(main *testing.T) {
	main.Parallel()

	err := &datatype.TypeError{Expected: datatype.Int}
	require.EqualError(main, err, "is not int")
}

func TestVariable_Type(main *testing.T) {
	main.Parallel()

	v, err := datatype.NewInt("42")
	require.NoError(main, err)
	assert.Equal(main, datatype.Int, v.Type())
}

func TestVariable_Bool(main *testing.T) {
	main.Parallel()

	main.Run("Ok", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewBool("true")
		require.NoError(t, err)

		b, err := v.Bool()
		require.NoError(t, err)
		assert.True(t, b)
	})

	main.Run("WrongType", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewInt("1")
		require.NoError(t, err)

		_, err = v.Bool()
		require.EqualError(t, err, "is not bool")
	})
}

func TestVariable_Int(main *testing.T) {
	main.Parallel()

	main.Run("Ok", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewInt("42")
		require.NoError(t, err)

		i, err := v.Int()
		require.NoError(t, err)
		assert.Equal(t, 42, i)
	})

	main.Run("WrongType", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewString("s")
		require.NoError(t, err)

		_, err = v.Int()
		require.EqualError(t, err, "is not int")
	})
}

func TestVariable_Float(main *testing.T) {
	main.Parallel()

	main.Run("Ok", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewFloat("3.14")
		require.NoError(t, err)

		f, err := v.Float()
		require.NoError(t, err)
		assert.InDelta(t, 3.14, f, 0.0001)
	})

	main.Run("WrongType", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewBool("true")
		require.NoError(t, err)

		_, err = v.Float()
		require.EqualError(t, err, "is not float")
	})
}

func TestVariable_String(main *testing.T) {
	main.Parallel()

	main.Run("Ok", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewString("hello")
		require.NoError(t, err)

		s, err := v.String()
		require.NoError(t, err)
		assert.Equal(t, "hello", s)
	})

	main.Run("WrongType", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewFloat("1.5")
		require.NoError(t, err)

		_, err = v.String()
		require.EqualError(t, err, "is not string")
	})
}

func TestNewBool(main *testing.T) {
	main.Parallel()

	main.Run("True", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewBool("true")
		require.NoError(t, err)
		assert.Equal(t, datatype.Bool, v.Type())

		b, err := v.Bool()
		require.NoError(t, err)
		assert.True(t, b)
	})

	main.Run("False", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewBool("false")
		require.NoError(t, err)

		b, err := v.Bool()
		require.NoError(t, err)
		assert.False(t, b)
	})

	main.Run("CaseInsensitiveWithWhitespace", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewBool("  TRUE  ")
		require.NoError(t, err)

		b, err := v.Bool()
		require.NoError(t, err)
		assert.True(t, b)
	})

	main.Run("Invalid", func(t *testing.T) {
		t.Parallel()

		_, err := datatype.NewBool("notabool")
		require.EqualError(t, err, "value: cannot be converted to a bool")
	})
}

func TestNewInt(main *testing.T) {
	main.Parallel()

	main.Run("Ok", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewInt("42")
		require.NoError(t, err)
		assert.Equal(t, datatype.Int, v.Type())

		i, err := v.Int()
		require.NoError(t, err)
		assert.Equal(t, 42, i)
	})

	main.Run("Negative", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewInt("-7")
		require.NoError(t, err)

		i, err := v.Int()
		require.NoError(t, err)
		assert.Equal(t, -7, i)
	})

	main.Run("WithWhitespace", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewInt("  42  ")
		require.NoError(t, err)

		i, err := v.Int()
		require.NoError(t, err)
		assert.Equal(t, 42, i)
	})

	main.Run("Invalid", func(t *testing.T) {
		t.Parallel()

		_, err := datatype.NewInt("notanint")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be converted to an int")
	})

	main.Run("Float", func(t *testing.T) {
		t.Parallel()

		_, err := datatype.NewInt("3.14")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be converted to an int")
	})
}

func TestNewFloat(main *testing.T) {
	main.Parallel()

	main.Run("Ok", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewFloat("3.14")
		require.NoError(t, err)
		assert.Equal(t, datatype.Float, v.Type())

		f, err := v.Float()
		require.NoError(t, err)
		assert.InDelta(t, 3.14, f, 0.0001)
	})

	main.Run("Integer", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewFloat("42")
		require.NoError(t, err)

		f, err := v.Float()
		require.NoError(t, err)
		assert.InDelta(t, 42.0, f, 0.0001)
	})

	main.Run("WithWhitespace", func(t *testing.T) {
		t.Parallel()

		v, err := datatype.NewFloat("  3.14  ")
		require.NoError(t, err)

		f, err := v.Float()
		require.NoError(t, err)
		assert.InDelta(t, 3.14, f, 0.0001)
	})

	main.Run("Invalid", func(t *testing.T) {
		t.Parallel()

		_, err := datatype.NewFloat("notafloat")
		require.Error(t, err)
		assert.Contains(t, err.Error(), "cannot be converted to a float")
	})
}

func TestNewString(main *testing.T) {
	main.Parallel()

	main.Run("Ok", func(t *testing.T) {
		t.Parallel()

		v := datatype.NewString("hello world")
		assert.Equal(t, datatype.String, v.Type())

		s, err := v.String()
		require.NoError(t, err)
		assert.Equal(t, "hello world", s)
	})

	main.Run("Empty", func(t *testing.T) {
		t.Parallel()

		v := datatype.NewString("")

		s, err := v.String()
		require.NoError(t, err)
		assert.Empty(t, s)
	})
}
