package jsengine

import (
	"fmt"
	"io"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/datatype"
	"github.com/dop251/goja"
	"github.com/rs/zerolog"
)

type Opts struct {
	code string
}

type JS struct {
	opts    Opts
	taskLog io.WriteCloser
	l       zerolog.Logger
}

func New(opts datatype.VarMap, taskLog io.WriteCloser, l zerolog.Logger) (*JS, error) {
	o, err := ValidateOpts(opts)
	if err != nil {
		return nil, fmt.Errorf("validate options: %w", err)
	}

	return &JS{
		opts:    o,
		taskLog: taskLog,
		l:       l,
	}, nil
}

func (j *JS) Execute(args datatype.VarMap) (datatype.VarMap, error) {
	defer func() {
		_ = j.taskLog.Close()
	}()

	if err := ValidateArgs(args); err != nil {
		return nil, fmt.Errorf("validate args: %w", err)
	}

	vm := goja.New()
	vm.SetFieldNameMapper(goja.UncapFieldNameMapper())

	if err := vm.Set("console", &console{w: j.taskLog}); err != nil {
		return nil, fmt.Errorf("set console: %w", err)
	}

	if err := vm.Set("_args", args.AsMap()); err != nil {
		return nil, fmt.Errorf("set _args: %w", err)
	}

	jsRes := make(map[string]any)
	if err := vm.Set("_res", jsRes); err != nil {
		return nil, fmt.Errorf("set _res: %w", err)
	}

	if _, err := vm.RunString(j.opts.code); err != nil {
		return nil, err
	}

	res := datatype.VarMap{}
	if err := res.FromMap(jsRes); err != nil {
		return nil, fmt.Errorf("process result: %w", err)
	}

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
