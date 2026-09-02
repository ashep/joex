package scheduler

import (
	"context"
	"fmt"
	"time"

	"github.com/ashep/joex/internal/job"
	"github.com/ashep/joex/internal/pipeline"
	"github.com/ashep/joex/internal/status"
	"github.com/rs/zerolog"
)

type pipelineService interface {
	FindByID(ctx context.Context, id string) (pipeline.Pipeline, error)
}

type jobService interface {
	FindJobByID(ctx context.Context, pipeID, jobID string) (job.Job, error)
	FindJobsByStatus(ctx context.Context, st []status.ProcessingStatus, limit int) ([]job.Job, error)
	ScheduleNextTask(ctx context.Context, pipeID string, jobID string) (job.Task, error)
	FindTasksByStatus(ctx context.Context, status status.ProcessingStatus, limit int) ([]job.Task, error)
}

type Scheduler struct {
	pipeSvc      pipelineService
	jobSvc       jobService
	pollInterval time.Duration
	now          func() time.Time
	l            zerolog.Logger
	stopped      chan struct{}
}

func New(
	pipeSvc pipelineService,
	jobSvc jobService,
	pollInterval time.Duration,
	now func() time.Time,
	l zerolog.Logger,
) *Scheduler {
	return &Scheduler{
		pipeSvc:      pipeSvc,
		jobSvc:       jobSvc,
		pollInterval: pollInterval,
		now:          now,
		l:            l,
		stopped:      make(chan struct{}),
	}
}

func (s *Scheduler) Run(ctx context.Context) error {
	sleep := time.Duration(0)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(sleep):
		}

		if err := s.run(ctx); err != nil {
			s.l.Error().Err(err).Send()
		}

		sleep = s.pollInterval
	}
}

func (s *Scheduler) run(ctx context.Context) error {
	jobs, err := s.jobSvc.FindJobsByStatus(ctx, []status.ProcessingStatus{status.New, status.Idle}, 10)
	if err != nil {
		return fmt.Errorf("find jobs by status: %w", err)
	}

	for _, j := range jobs {
		t, err := s.jobSvc.ScheduleNextTask(ctx, j.PipelineID, j.ID)
		if err != nil {
			s.l.Error().Err(err).
				Str("pipeline_id", j.PipelineID).
				Str("job_id", j.ID).
				Int("job_v", j.Version).
				Msg("failed to schedule task")
			continue
		}

		s.l.Info().
			Str("pipeline_id", j.PipelineID).
			Str("job_id", j.ID).
			Int("job_v", j.Version).
			Str("task_id", t.ID).
			Int("step_id", t.StepID).
			Msg("task scheduled")
	}

	return nil
}

func (s *Scheduler) Wait() error {
	<-s.stopped
	return nil
}
