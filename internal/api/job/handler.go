package job

import (
	"context"
	"time"

	"github.com/ashep/joex/internal/datatype"
	"github.com/ashep/joex/internal/job"
	"github.com/rs/zerolog"
)

type jobService interface {
	CreateJob(ctx context.Context, pipeID string, args datatype.VarMap) (job.Job, error)
}

type Handler struct {
	jobSvc jobService
	now    func() time.Time
	l      zerolog.Logger
}

func New(pipSvc jobService, now func() time.Time, l zerolog.Logger) *Handler {
	return &Handler{
		jobSvc: pipSvc,
		now:    now,
		l:      l,
	}
}
