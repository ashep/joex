package auth

import (
	"context"
	"errors"
	"strings"

	"connectrpc.com/connect"
)

type apiKeyCtxKey struct{}

type APIKeyInterceptor struct {
	defaultAPIKey string
	apiKeys       map[string]string
}

func NewAPIKeyInterceptor(apiKeys map[string]string) *APIKeyInterceptor {
	if len(apiKeys) == 0 {
		panic(errors.New("at least one API key is required"))
	}
	if _, ok := apiKeys["default"]; !ok {
		panic(errors.New("default API key is required"))
	}

	m := make(map[string]string)
	for k, v := range apiKeys {
		m[v] = k // map keys are api key values, not names
	}

	return &APIKeyInterceptor{
		defaultAPIKey: m["default"],
		apiKeys:       m,
	}
}

func (i *APIKeyInterceptor) WrapUnary(next connect.UnaryFunc) connect.UnaryFunc {
	return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
		if req.Spec().IsClient {
			req.Header().Set("Authorization", "Bearer "+i.defaultAPIKey)
			return next(ctx, req)
		}

		bIdx := strings.Index(req.Header().Get("Authorization"), "Bearer ")
		if bIdx == -1 {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
		}

		tok := req.Header().Get("Authorization")[7:]
		if _, ok := i.apiKeys[tok]; !ok {
			return nil, connect.NewError(connect.CodeUnauthenticated, errors.New("unauthenticated"))
		}

		return next(context.WithValue(ctx, apiKeyCtxKey{}, i.apiKeys[tok]), req)
	}
}

func (i *APIKeyInterceptor) WrapStreamingClient(next connect.StreamingClientFunc) connect.StreamingClientFunc {
	return func(ctx context.Context, spec connect.Spec) connect.StreamingClientConn {
		// TODO: implement
		return next(ctx, spec)
	}
}

func (i *APIKeyInterceptor) WrapStreamingHandler(next connect.StreamingHandlerFunc) connect.StreamingHandlerFunc {
	return func(ctx context.Context, conn connect.StreamingHandlerConn) error {
		// TODO: implement
		return next(ctx, conn)
	}
}
