package engine

import (
	"errors"

	"github.com/ashep/joex/internal/datatype"
	"github.com/ashep/joex/internal/engine/jsengine"
	"github.com/rs/zerolog"
)

var UnsupportedEngineTypeErr = errors.New("unsupported engine type")

const (
	Unknown Type = "unknown"
	JS      Type = "js"
)

type Engine interface {
	Execute(args datatype.VarMap) (datatype.VarMap, error)
}

type Type string

func ValidateType(e Type) error {
	switch e {
	case JS:
		return nil
	default:
		return UnsupportedEngineTypeErr
	}
}

func ValidateOpts(e Type, opts datatype.VarMap) error {
	switch e {
	case JS:
		_, err := jsengine.ValidateOpts(opts)
		return err
	default:
		return UnsupportedEngineTypeErr
	}
}

func ValidateArgs(e Type, args datatype.VarMap) error {
	switch e {
	case JS:
		return jsengine.ValidateArgs(args)
	default:
		return UnsupportedEngineTypeErr
	}
}

func New(e Type, opts datatype.VarMap, l zerolog.Logger) (Engine, error) {
	switch e {
	case JS:
		return jsengine.New(opts, l.With().Str("engine", string(e)).Logger())
	default:
		return nil, UnsupportedEngineTypeErr
	}
}
