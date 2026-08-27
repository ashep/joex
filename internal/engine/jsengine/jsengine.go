package jsengine

import (
	"errors"
	"fmt"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/datatype"
	"github.com/rs/zerolog"
)

type Opts struct {
	code string
}

type JS struct {
	opts Opts
	l    zerolog.Logger
}

func New(opts datatype.VarMap, l zerolog.Logger) (*JS, error) {
	o, err := ValidateOpts(opts)
	if err != nil {
		return nil, fmt.Errorf("validate options: %w", err)
	}

	return &JS{
		opts: o,
		l:    l,
	}, nil
}

func (j *JS) Execute(args datatype.VarMap) (datatype.VarMap, error) {
	if err := ValidateArgs(args); err != nil {
		return nil, fmt.Errorf("validate args: %w", err)
	}

	foo, err := args.Int("foo")
	if err != nil {
		return nil, fmt.Errorf("get foo: %w", err)
	}

	if foo == 123 {
		return nil, errors.New("synthetic error")
	}

	res := datatype.VarMap{}
	_ = res.SetInt("foo", "321")

	return res, nil
}

func ValidateOpts(opts datatype.VarMap) (Opts, error) {
	code, err := opts.String("code")
	if err != nil {
		return Opts{}, err
	}

	if code == "" {
		return Opts{}, apperr.NewRequiredArg("code")
	}

	return Opts{code: code}, nil
}

func ValidateArgs(_ datatype.VarMap) error {
	return nil
}
