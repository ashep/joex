package pipeline

import (
	"encoding/json"
	"fmt"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/status"
	"github.com/ashep/joex/pkg/typeutil"
)

type Pipeline struct {
	ID        string               `json:"id" dynamodbav:"id"`
	Status    status.EnabledStatus `json:"status" dynamodbav:"status"`
	Steps     []Step               `json:"steps" dynamodbav:"steps"`
	Version   int                  `json:"version" dynamodbav:"version"`
	CreatedAt typeutil.UnixTimeMs  `json:"created_at" dynamodbav:"created_at"`
	UpdatedAt typeutil.UnixTimeMs  `json:"updated_at" dynamodbav:"updated_at"`
}

func (p Pipeline) Validate() error {
	if p.ID == "" {
		return apperr.NewRequiredArg("id")
	}

	if err := status.ValidateEnabledStatus(p.Status); err != nil {
		return err
	}

	if len(p.Steps) == 0 {
		return apperr.NewInvalidArg("steps", "must have at least one step")
	}

	for i, step := range p.Steps {
		if err := step.Validate(); err != nil {
			return fmt.Errorf("steps[%d]: %w", i, err)
		}
	}

	if p.CreatedAt.IsZero() {
		return apperr.NewRequiredArg("created_at")
	}

	if p.UpdatedAt.IsZero() {
		return apperr.NewRequiredArg("updated_at")
	}

	return nil
}

func (p Pipeline) Step(id int) (Step, error) {
	if id >= len(p.Steps) || id < 0 {
		return Step{}, StepNotFoundError
	}

	return p.Steps[id], nil
}

func (p Pipeline) StreamEventKey() string {
	return p.ID
}

func (p Pipeline) StreamEventValue() []byte {
	b, err := json.Marshal(p)
	if err != nil {
		panic(err)
	}

	return b
}

func (p Pipeline) StreamEventHeaders() [][2]string {
	return nil
}
