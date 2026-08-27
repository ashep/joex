package app

import (
	"fmt"
	"maps"
	"time"

	"connectrpc.com/connect"
	"connectrpc.com/validate"
	"github.com/ashep/go-app/httpserver"
	"github.com/ashep/go-app/runner"
	"github.com/ashep/go-app/taskrunner"
	jobhandler "github.com/ashep/joex/internal/api/job"
	pipelinehandler "github.com/ashep/joex/internal/api/pipeline"
	"github.com/ashep/joex/internal/executor"
	"github.com/ashep/joex/internal/job"
	"github.com/ashep/joex/internal/pipeline"
	"github.com/ashep/joex/internal/scheduler"
	"github.com/ashep/joex/pkg/auth"
	joexconn "github.com/ashep/joex/sdk/proto/joex/v1/joexv1connect"
	awscfg "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
)

func Run(rt *runner.Runtime[Config]) error {
	cfg := rt.Cfg
	l := rt.Log

	if cfg.Now == nil {
		cfg.Now = time.Now
	}

	awsCfg, err := awscfg.LoadDefaultConfig(rt.Ctx)
	if err != nil {
		return fmt.Errorf("load aws config: %w", err)
	}

	tr := taskrunner.New(taskrunner.WithLogger(l), taskrunner.WithPanic(rt.Debug))

	ddb := dynamodb.NewFromConfig(awsCfg)

	icps := connect.WithInterceptors(
		auth.NewAPIKeyInterceptor(apiKeysFromConfig(cfg)),
		validate.NewInterceptor(validate.WithoutErrorDetails()),
	)

	pipeRepo, err := pipeline.NewRepo(ddb, cfg.Database.DDB.Table, l.With().Str("pkg", "pipeline_repo").Logger())
	if err != nil {
		return fmt.Errorf("new pipeline repo: %w", err)
	}
	pipeSvc := pipeline.NewService(pipeRepo)

	jobRepo, err := job.NewRepo(ddb, cfg.Database.DDB.Table, cfg.Now, l)
	if err != nil {
		return fmt.Errorf("new job repo: %w", err)
	}
	jobSvc := job.NewService(pipeSvc, jobRepo, cfg.Now, l)

	if err := startTaskScheduler(rt, tr, pipeSvc, jobSvc); err != nil {
		return err
	}

	if err := startTaskExecutor(rt, tr, jobSvc); err != nil {
		return err
	}

	srv := httpserver.New(
		httpserver.WithAddr(cfg.Server.Addr),
		httpserver.WithHTTP1(true),
		httpserver.WithUnencryptedHTTP2(true),
	)
	srv.Handle(joexconn.NewPipelineServiceHandler(pipelinehandler.New(pipeSvc, cfg.Now, l), icps))
	srv.Handle(joexconn.NewJobServiceHandler(jobhandler.New(jobSvc, cfg.Now, l), icps))

	l.Info().Str("addr", srv.Listener().Addr().String()).Msg("server addr")
	if err := tr.Run(rt.Ctx, "server", srv); err != nil {
		return fmt.Errorf("start server: %w", err)
	}

	return tr.Wait(rt.Ctx)
}

func startTaskScheduler(rt *runner.Runtime[Config], tr *taskrunner.Runner, pipeSvc *pipeline.Service, jobSvc *job.Service) error {
	l := rt.Log.With().Str("pkg", "task_scheduler").Logger()
	taskSched := scheduler.New(pipeSvc, jobSvc, rt.Cfg.Now, l)

	if err := tr.Run(rt.Ctx, "task scheduler", taskSched); err != nil {
		return fmt.Errorf("start task scheduler: %w", err)
	}

	return nil
}

func startTaskExecutor(rt *runner.Runtime[Config], tr *taskrunner.Runner, jobSvc *job.Service) error {
	l := rt.Log

	ex := executor.New(jobSvc, l.With().Str("pkg", "task_executor").Logger())
	if err := tr.Run(rt.Ctx, "task executor", ex); err != nil {
		return fmt.Errorf("run task executor: %w", err)
	}

	return nil
}

func apiKeysFromConfig(cfg *Config) map[string]string {
	apiKeys := make(map[string]string)
	apiKeys["default"] = cfg.Server.Auth.APIKey
	maps.Copy(apiKeys, cfg.Server.Auth.APIKeys)
	return apiKeys
}
