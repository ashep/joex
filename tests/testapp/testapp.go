package testapp

import (
	"net/http"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/ashep/go-app/testlogger"
	"github.com/ashep/go-app/testrunner"
	"github.com/ashep/joex/internal/app"
	"github.com/ashep/joex/pkg/auth"
	"github.com/ashep/joex/sdk/proto/joex/v1/joexv1connect"
)

type tRunner interface {
	Logger() *testlogger.Logger
}

type TestApp struct {
	Cfg app.Config

	t   *testing.T
	rnr tRunner
}

type ConfigOption func(*app.Config)

func New(t *testing.T, opts ...ConfigOption) *TestApp {
	t.Helper()

	cfg := app.Config{
		Database: app.DatabaseConfig{
			DDB: app.DatabaseDDBConfig{
				Table: "test",
			},
		},
		Server: app.ServerConfig{
			Addr: testrunner.RandLocalTCPAddr(t).String(),
			Auth: app.ServerAuthConfig{
				APIKey: "default",
			},
		},
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	rnr := testrunner.New(t, app.Run, cfg).
		SetHTTPReadyStartWaiter("http://"+cfg.Server.Addr+"/metrics", time.Second*5).
		Start()

	ta := &TestApp{
		t:   t,
		Cfg: cfg,
		rnr: rnr,
	}

	return ta
}

func (ta *TestApp) AssertNoWarnsAndErrors() {
	ta.rnr.Logger().AssertNoWarnsAndErrors()
}

func (ta *TestApp) NewJobServiceClient(apiKey string) joexv1connect.JobServiceClient {
	icps := connect.WithInterceptors(auth.NewAPIKeyInterceptor(map[string]string{
		"default": apiKey,
	}))

	return joexv1connect.NewJobServiceClient(http.DefaultClient, "http://"+ta.Cfg.Server.Addr, icps)
}
