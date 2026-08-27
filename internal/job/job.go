package job

import (
	"encoding/json"
	"time"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/datatype"
	"github.com/ashep/joex/internal/status"
	"github.com/ashep/joex/pkg/typeutil"
)

type Job struct {
	ID             string                  `json:"id" dynamodbav:"id"`
	PipelineID     string                  `json:"pipeline_id" dynamodbav:"pipeline_id"`
	CurStep        int                     `json:"cur_step" dynamodbav:"cur_step"`
	MaxStep        int                     `json:"max_step" dynamodbav:"max_step"`
	Args           datatype.VarMap         `json:"args" dynamodbav:"args"`
	Status         status.ProcessingStatus `json:"status" dynamodbav:"status"`
	LastTaskResult datatype.VarMap         `json:"lt_result" dynamodbav:"lt_result"`
	Version        int                     `json:"version" dynamodbav:"version"`
	CreatedAt      typeutil.UnixTimeMs     `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt      typeutil.UnixTimeMs     `json:"updated_at" dynamodbav:"updated_at"`
}

func (j Job) StreamEventKey() string {
	return j.ID
}

func (j Job) StreamEventValue() []byte {
	b, _ := json.Marshal(j)
	return b
}

func (j Job) StreamEventHeaders() [][2]string {
	return nil
}

func (j Job) StreamEventCreatedAt() time.Time {
	return j.CreatedAt.AsTime()
}

func (j Job) Validate() error {
	if j.ID == "" {
		return apperr.NewRequiredArg("id")
	}

	if j.PipelineID == "" {
		return apperr.NewRequiredArg("pipeline_id")
	}

	if j.CurStep < 0 {
		return apperr.NewInvalidArg("cur_step", "must not be negative")
	}

	if j.MaxStep < 0 {
		return apperr.NewInvalidArg("max_step", "must not be negative")
	}

	if err := status.ValidateProcessingStatus(j.Status); err != nil {
		return err
	}

	if j.CreatedAt.IsZero() {
		return apperr.NewRequiredArg("created_at")
	}

	if j.UpdatedAt.IsZero() {
		return apperr.NewRequiredArg("updated_at")
	}

	return nil
}
