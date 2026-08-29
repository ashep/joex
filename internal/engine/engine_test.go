package engine_test

import (
	"testing"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/datatype"
	"github.com/ashep/joex/internal/engine"
	"github.com/ashep/joex/internal/engine/jsengine"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// unsupportedTypes are the engine types no branch of the dispatch switches accepts.
func unsupportedTypes() []engine.Type {
	return []engine.Type{engine.Unspecified, "", "aBogusEngine"}
}

// requireUnsupportedEngineError asserts err is the invalid-arg error the API layer
// relies on to map an unsupported engine onto an invalid_argument response.
func requireUnsupportedEngineError(t *testing.T, err error, typ engine.Type) {
	t.Helper()

	require.Error(t, err)

	var argErr apperr.InvalidArgError
	require.ErrorAs(t, err, &argErr, "it is not an invalid arg error")
	assert.Equal(t, string(typ), argErr.Arg)
	assert.Equal(t, "unsupported engine", argErr.Reason)
}

// validJSOpts builds opts the JS engine accepts. The unsupported-engine cases use
// them too, to show those are rejected on the engine type alone, not on the opts.
func validJSOpts(t *testing.T) datatype.VarMap {
	t.Helper()

	opts := datatype.VarMap{}
	require.NoError(t, opts.SetString("code", "1+1"))

	return opts
}

func TestValidateType(main *testing.T) {
	main.Parallel()

	main.Run("Unsupported", func(t *testing.T) {
		t.Parallel()

		for _, typ := range unsupportedTypes() {
			requireUnsupportedEngineError(t, engine.ValidateType(typ), typ)
		}
	})

	main.Run("JS", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, engine.ValidateType(engine.JS))
	})

	main.Run("UnsupportedErrorMessage", func(t *testing.T) {
		t.Parallel()

		assert.EqualError(t, engine.ValidateType(engine.Unspecified), "unspecified: unsupported engine")
		assert.EqualError(t, engine.ValidateType("aBogusEngine"), "aBogusEngine: unsupported engine")

		// An empty type carries no name, and apperr renders such an error as an empty
		// string. The error is still a non-nil invalid arg error, so the API layer maps
		// it to invalid_argument, but the client gets a blank message.
		assert.EqualError(t, engine.ValidateType(""), "")
	})
}

func TestValidateOpts(main *testing.T) {
	main.Parallel()

	main.Run("Unsupported", func(t *testing.T) {
		t.Parallel()

		for _, typ := range unsupportedTypes() {
			requireUnsupportedEngineError(t, engine.ValidateOpts(typ, validJSOpts(t)), typ)
		}
	})

	main.Run("JS", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, engine.ValidateOpts(engine.JS, validJSOpts(t)))
	})

	main.Run("JSReturnsEngineErrorAsIs", func(t *testing.T) {
		t.Parallel()

		opts := datatype.VarMap{}

		_, expErr := jsengine.ValidateOpts(opts)
		require.Error(t, expErr, "opts are expected to be rejected by the engine")

		assert.Equal(t, expErr, engine.ValidateOpts(engine.JS, opts))
	})
}

func TestValidateArgs(main *testing.T) {
	main.Parallel()

	main.Run("Unsupported", func(t *testing.T) {
		t.Parallel()

		for _, typ := range unsupportedTypes() {
			requireUnsupportedEngineError(t, engine.ValidateArgs(typ, datatype.VarMap{}), typ)
		}
	})

	main.Run("JS", func(t *testing.T) {
		t.Parallel()

		require.NoError(t, engine.ValidateArgs(engine.JS, datatype.VarMap{}))
	})
}

func TestNew(main *testing.T) {
	main.Parallel()

	main.Run("Unsupported", func(t *testing.T) {
		t.Parallel()

		for _, typ := range unsupportedTypes() {
			e, err := engine.New(typ, validJSOpts(t), zerolog.Nop())

			requireUnsupportedEngineError(t, err, typ)
			assert.Nil(t, e)
		}
	})

	main.Run("JS", func(t *testing.T) {
		t.Parallel()

		e, err := engine.New(engine.JS, validJSOpts(t), zerolog.Nop())

		require.NoError(t, err)
		require.NotNil(t, e)
	})

	main.Run("JSInvalidOpts", func(t *testing.T) {
		t.Parallel()

		e, err := engine.New(engine.JS, datatype.VarMap{}, zerolog.Nop())

		require.Error(t, err)
		assert.Nil(t, e)

		var argErr apperr.InvalidArgError
		assert.NotErrorAs(t, err, &argErr, "engine opts errors must not be reported as an unsupported engine")
	})
}
