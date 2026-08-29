package connecterr

import (
	"errors"

	"connectrpc.com/connect"
	apperr "github.com/ashep/go-app/errors"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type errCodeGen func() string

var ErrorCodeGenerator errCodeGen = uuid.NewString

func New(err error, l zerolog.Logger) *connect.Error {
	if err == nil {
		return nil
	}

	if _, ok := errors.AsType[apperr.InvalidArgError](err); ok {
		return connect.NewError(connect.CodeInvalidArgument, err)
	}

	if _, ok := errors.AsType[apperr.NotFoundError](err); ok {
		return connect.NewError(connect.CodeNotFound, err)
	}

	code := ErrorCodeGenerator()
	l.Error().Err(err).Str("err_code", code).Msg("internal error")

	return connect.NewError(connect.CodeInternal, errors.New("Error "+code))
}
