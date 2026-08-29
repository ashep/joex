package bufassert

import (
	"testing"

	"connectrpc.com/connect"
	"github.com/stretchr/testify/require"
)

func RequireError(t *testing.T, err error, expCode connect.Code, expMsg string) {
	require.Error(t, err)

	var cErr *connect.Error
	require.ErrorAs(t, err, &cErr, "it is not a connect error")
	require.Equalf(t, expCode, cErr.Code(), "error code is not %s, but %s", expCode.String(), cErr.Code().String())
	require.Equal(t, expMsg, cErr.Message(), "error message does not match")
}

func RequireInvalidArgumentError(t *testing.T, err error, expMsg string) {
	RequireError(t, err, connect.CodeInvalidArgument, expMsg)
}

func RequireInternalError(t *testing.T, err error, expMsg string) {
	RequireError(t, err, connect.CodeInternal, expMsg)
}
