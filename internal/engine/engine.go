package engine

import (
	"io"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/datatype"
	"github.com/ashep/joex/internal/engine/jsengine"
	"github.com/rs/zerolog"
)

const (
	Unspecified Type = "unspecified"
	JS          Type = "js"
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
		return apperr.NewInvalidArg(string(e), "unsupported engine")
	}
}

func ValidateOpts(e Type, opts datatype.VarMap) error {
	switch e {
	case JS:
		_, err := jsengine.ValidateOpts(opts)
		return err
	default:
		return apperr.NewInvalidArg(string(e), "unsupported engine")
	}
}

func ValidateArgs(e Type, args datatype.VarMap) error {
	switch e {
	case JS:
		return jsengine.ValidateArgs(args)
	default:
		return apperr.NewInvalidArg(string(e), "unsupported engine")
	}
}

func New(e Type, opts datatype.VarMap, taskLog io.WriteCloser, l zerolog.Logger) (Engine, error) {
	switch e {
	case JS:
		return jsengine.New(opts, taskLog, l.With().Str("engine", string(e)).Logger())
	default:
		return nil, apperr.NewInvalidArg(string(e), "unsupported engine")
	}
}
