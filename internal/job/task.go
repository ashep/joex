package job

import (
	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/datatype"
	"github.com/ashep/joex/internal/engine"
	"github.com/ashep/joex/internal/status"
	"github.com/ashep/joex/pkg/typeutil"
)

type Task struct {
	ID         string                  `json:"id" dynamodbav:"id"`
	JobID      string                  `json:"job_id" dynamodbav:"job_id"`
	PipelineID string                  `json:"pipeline_id" dynamodbav:"pipeline_id"`
	StepID     int                     `json:"step_id" dynamodbav:"step_id"`
	Engine     engine.Type             `json:"engine" dynamodbav:"engine"`
	Opts       datatype.VarMap         `json:"opts" dynamodbav:"opts"`
	Args       datatype.VarMap         `json:"args" dynamodbav:"args"`
	AllowFail  bool                    `json:"allow_fail" dynamodbav:"allow_fail"`
	IsLast     bool                    `json:"is_last" dynamodbav:"is_last"`
	Status     status.ProcessingStatus `json:"status" dynamodbav:"status"`
	Result     datatype.VarMap         `json:"result" dynamodbav:"result"`
	Log        string                  `json:"log" dynamodbav:"log"`
	CreatedAt  typeutil.UnixTimeMs     `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt  typeutil.UnixTimeMs     `json:"updated_at" dynamodbav:"updated_at"`
	Version    int                     `json:"version" dynamodbav:"version"`
}

func (t Task) Validate() error {
	if t.ID == "" {
		return apperr.NewRequiredArg("id")
	}

	if t.JobID == "" {
		return apperr.NewRequiredArg("job_id")
	}

	if t.PipelineID == "" {
		return apperr.NewRequiredArg("pipeline_id")
	}

	if t.StepID < 0 {
		return apperr.NewInvalidArg("step_id", "must not be negative")
	}

	if err := engine.ValidateType(t.Engine); err != nil {
		return apperr.NewInvalidArg("engine", err.Error())
	}

	if err := engine.ValidateOpts(t.Engine, t.Opts); err != nil {
		return apperr.NewInvalidArg("opts", err.Error())
	}

	if err := engine.ValidateArgs(t.Engine, t.Args); err != nil {
		return apperr.NewInvalidArg("args", err.Error())
	}

	if err := status.ValidateProcessingStatus(t.Status); err != nil {
		return err
	}

	if t.CreatedAt.IsZero() {
		return apperr.NewRequiredArg("created_at")
	}

	if t.UpdatedAt.IsZero() {
		return apperr.NewRequiredArg("updated_at")
	}

	return nil
}
