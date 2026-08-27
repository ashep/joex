package job

import (
	"context"
	"fmt"
	"time"

	"github.com/ashep/joex/internal/datatype"
	"github.com/ashep/joex/internal/pipeline"
	"github.com/ashep/joex/internal/status"
	"github.com/ashep/joex/pkg/ddbtx"
	"github.com/ashep/joex/pkg/typeutil"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

type pipelineService interface {
	FindByID(ctx context.Context, id string) (pipeline.Pipeline, error)
}

type repo interface {
	NewTx() *ddbtx.Tx

	CreateJob(tx *ddbtx.Tx, job Job) error
	FindJobByID(ctx context.Context, pipeID, jobID string) (Job, error)
	FindJobsByStatus(ctx context.Context, statuses []status.ProcessingStatus, limit int) ([]Job, error)
	UpdateJob(tx *ddbtx.Tx, j Job) error

	CreateTask(tx *ddbtx.Tx, task Task) error
	UpdateTask(tx *ddbtx.Tx, t Task) error
	FindTaskByID(ctx context.Context, jobID, taskID string) (Task, error)
	FindTasksByStatus(ctx context.Context, status status.ProcessingStatus, limit int) ([]Task, error)
}

type Service struct {
	pipeSvc pipelineService
	repo    repo
	now     func() time.Time
	l       zerolog.Logger
}

func NewService(pipeSvc pipelineService, repo repo, now func() time.Time, l zerolog.Logger) *Service {
	return &Service{
		pipeSvc: pipeSvc,
		repo:    repo,
		now:     now,
		l:       l,
	}
}

func (s *Service) CreateJob(ctx context.Context, pipeID string, args datatype.VarMap) (Job, error) {
	now := s.now()

	p, err := s.pipeSvc.FindByID(ctx, pipeID)
	if err != nil {
		return Job{}, fmt.Errorf("find pipeline: %w", err)
	}

	j := Job{
		ID:         uuid.NewString(),
		PipelineID: p.ID,
		CurStep:    0,
		MaxStep:    len(p.Steps) - 1,
		Args:       args,
		Status:     status.New,
		Version:    0,
		CreatedAt:  typeutil.UnixTimeMs(now),
		UpdatedAt:  typeutil.UnixTimeMs(now),
	}
	if err := j.Validate(); err != nil {
		return Job{}, fmt.Errorf("new job: %w", err)
	}

	tx := s.repo.NewTx()

	if err := s.repo.CreateJob(tx, j); err != nil {
		return Job{}, fmt.Errorf("repo: create job: %w", err)
	}

	if _, err := tx.Commit(ctx); err != nil {
		return Job{}, fmt.Errorf("repo: commit: %w", err)
	}

	return j, nil
}

func (s *Service) FindJobByID(ctx context.Context, pipeID, jobID string) (Job, error) {
	return s.repo.FindJobByID(ctx, pipeID, jobID)
}

func (s *Service) FindJobsByStatus(ctx context.Context, st []status.ProcessingStatus, limit int) ([]Job, error) {
	return s.repo.FindJobsByStatus(ctx, st, limit)
}

// ScheduleNextTask schedules a next task of a job for execution.
// The job must be new or idle.
func (s *Service) ScheduleNextTask(ctx context.Context, pipeID string, jobID string) (Task, error) {
	job, err := s.repo.FindJobByID(ctx, pipeID, jobID)
	if err != nil {
		return Task{}, fmt.Errorf("find job: %w", err)
	}

	if job.Status != status.New && job.Status != status.Idle {
		return Task{}, fmt.Errorf("job status is %s", job.Status)
	}

	job.Status = status.Running

	p, err := s.pipeSvc.FindByID(ctx, job.PipelineID)
	if err != nil {
		return Task{}, fmt.Errorf("find pipeline: %w", err)
	}

	step, err := p.Step(job.CurStep)
	if err != nil {
		return Task{}, fmt.Errorf("get job step: %w", err)
	}

	args := job.Args
	if step.ID > 0 {
		args = job.LastTaskResult
	}

	now := s.now()
	t := Task{
		ID:         uuid.NewString(),
		JobID:      job.ID,
		PipelineID: job.PipelineID,
		StepID:     step.ID,
		Engine:     step.Engine,
		Opts:       step.Opts,
		Args:       args,
		AllowFail:  step.AllowFail,
		IsLast:     step.ID == len(p.Steps)-1,
		Status:     status.New,
		CreatedAt:  typeutil.UnixTimeMs(now),
		UpdatedAt:  typeutil.UnixTimeMs(now),
	}
	if err := t.Validate(); err != nil {
		return Task{}, fmt.Errorf("new task: %w", err)
	}

	tx := s.repo.NewTx()
	if err := s.repo.CreateTask(tx, t); err != nil {
		return Task{}, fmt.Errorf("repo: create task: %w", err)
	}
	if err := s.repo.UpdateJob(tx, job); err != nil {
		return Task{}, fmt.Errorf("repo: update job: %w", err)
	}
	if _, err := tx.Commit(ctx); err != nil {
		return Task{}, fmt.Errorf("repo: commit: %w", err)
	}

	return t, nil
}

func (s *Service) FindTaskByID(ctx context.Context, jobID, taskID string) (Task, error) {
	return s.repo.FindTaskByID(ctx, jobID, taskID)
}

func (s *Service) FindTasksByStatus(ctx context.Context, status status.ProcessingStatus, limit int) ([]Task, error) {
	return s.repo.FindTasksByStatus(ctx, status, limit)
}

func (s *Service) CompleteTask(
	ctx context.Context,
	jobID string,
	taskID string,
	res datatype.VarMap,
	resErr error,
) error {
	t, err := s.FindTaskByID(ctx, jobID, taskID)
	if err != nil {
		return fmt.Errorf("find task: %w", err)
	}

	if resErr != nil {
		if res == nil {
			res = make(datatype.VarMap)
		}
		if err := res.SetString("error", resErr.Error()); err != nil {
			return fmt.Errorf("set result error: %w", err)
		}
		t.Status = status.Failed
	} else {
		t.Status = status.Done
	}

	t.Result = res
	t.UpdatedAt = typeutil.UnixTimeMs(s.now())

	j, err := s.FindJobByID(ctx, t.PipelineID, jobID)
	if err != nil {
		return fmt.Errorf("find job: %w", err)
	}

	j.LastTaskResult = t.Result

	if j.CurStep == j.MaxStep { // last task in the job
		j.Status = t.Status
		if t.AllowFail {
			j.Status = status.Done
		}
	} else if resErr != nil && !t.AllowFail { // task is not last, but it causes fail of the entire job
		j.Status = status.Failed
	} else { // job is waiting for the next task to be scheduled
		j.CurStep++
		j.Status = status.Idle
	}

	tx := s.repo.NewTx()
	if err := s.repo.UpdateTask(tx, t); err != nil {
		return fmt.Errorf("repo: update task: %w", err)
	}
	if err := s.repo.UpdateJob(tx, j); err != nil {
		return fmt.Errorf("repo: update job: %w", err)
	}
	if _, err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("repo: commit: %w", err)
	}

	return nil
}
