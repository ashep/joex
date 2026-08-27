package connecterr

import (
	"errors"
	"fmt"
	"time"

	"connectrpc.com/connect"
	apperr "github.com/ashep/go-app/errors"
)

func New(err error, now func() time.Time) *connect.Error {
	if err == nil {
		return nil
	}

	if tErr, ok := errors.AsType[apperr.InvalidArgError](err); ok {
		return connect.NewError(connect.CodeInvalidArgument, tErr)
	}

	if tErr, ok := errors.AsType[apperr.NotFoundError](err); ok {
		return connect.NewError(connect.CodeNotFound, tErr)
	}

	return connect.NewError(connect.CodeInternal, fmt.Errorf("internal error: %d", now().UnixMilli()))
}
