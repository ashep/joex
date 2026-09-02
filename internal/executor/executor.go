package executor

import (
	"context"
	"fmt"
	"time"

	"github.com/ashep/joex/internal/datatype"
	"github.com/ashep/joex/internal/engine"
	"github.com/ashep/joex/internal/job"
	"github.com/ashep/joex/internal/status"
	"github.com/rs/zerolog"
)

type jobService interface {
	FindTasksByStatus(ctx context.Context, st status.ProcessingStatus, limit int) ([]job.Task, error)
	CompleteTask(ctx context.Context, jobID string, taskID string, res datatype.VarMap, resErr error, log fmt.Stringer) error
}

// Executor processes new tasks.
type Executor struct {
	jobSvc       jobService
	pollInterval time.Duration
	logBufSize   int
	l            zerolog.Logger
	stopped      chan struct{}
}

func New(jobSvc jobService, pollInterval time.Duration, logBufSize int, l zerolog.Logger) *Executor {
	return &Executor{
		jobSvc:       jobSvc,
		pollInterval: pollInterval,
		logBufSize:   logBufSize,
		l:            l,
		stopped:      make(chan struct{}),
	}
}

func (e *Executor) Run(ctx context.Context) error {
	defer close(e.stopped)

	sleep := time.Duration(0)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}

		sleep = e.pollInterval

		tasks, err := e.jobSvc.FindTasksByStatus(ctx, status.New, 10)
		if err != nil {
			e.l.Error().Err(err).Msg("failed to find tasks")
			continue
		}

		for _, t := range tasks {
			// TODO: spawn goroutine instead
			if err := e.execute(ctx, t); err != nil {
				e.l.Error().Err(err).
					Str("job_id", t.JobID).
					Str("task_id", t.ID).
					Int("task_ver", t.Version).
					Msg("failed to execute task")
			}
		}
	}
}

func (e *Executor) execute(ctx context.Context, t job.Task) error {
	if err := t.Validate(); err != nil {
		return fmt.Errorf("invalid task: %w", err)
	}

	tLog := newLogger(e.logBufSize)

	eng, err := engine.New(t.Engine, t.Opts, tLog, e.l)
	if err != nil {
		return fmt.Errorf("step %d: instantiate engine: %w", t.StepID, err)
	}

	e.l.Info().
		Str("pipe_id", t.PipelineID).
		Str("job_id", t.JobID).
		Str("task_id", t.ID).
		Int("step_id", t.StepID).
		Int("task_ver", t.Version).
		Msg("task started")

	res, exErr := eng.Execute(t.Args)
	if err := e.jobSvc.CompleteTask(ctx, t.JobID, t.ID, res, exErr, tLog); err != nil {
		return fmt.Errorf("job service: complete task: %w", err)
	}

	e.l.Info().
		Str("pipe_id", t.PipelineID).
		Str("job_id", t.JobID).
		Str("task_id", t.ID).
		Int("step_id", t.StepID).
		Int("task_ver", t.Version).
		Bool("task_err", exErr != nil).
		Msg("task executed")

	return nil
}

func (e *Executor) Wait() error {
	<-e.stopped
	return nil
}
