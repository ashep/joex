package pipeline

import (
	"errors"

	apperr "github.com/ashep/go-app/errors"
	"github.com/ashep/joex/internal/datatype"
	"github.com/ashep/joex/internal/engine"
)

var ErrStepNotFound = errors.New("step not found")

type Step struct {
	ID        int             `json:"id" dynamodbav:"id"`
	Engine    engine.Type     `json:"engine" dynamodbav:"engine"`
	Opts      datatype.VarMap `json:"opts" dynamodbav:"opts"`
	AllowFail bool            `json:"allow_fail" dynamodbav:"allow_fail"`
}

func MakeStep(id int, e engine.Type, opts datatype.VarMap, allowFail bool) (Step, error) {
	if id < 0 {
		return Step{}, apperr.NewInvalidArg("id", "must be greater or equal to zero")
	}

	t := Step{
		ID:        id,
		Engine:    e,
		Opts:      opts,
		AllowFail: allowFail,
	}

	return t, t.Validate()
}

func (s Step) Validate() error {
	if err := engine.ValidateOpts(s.Engine, s.Opts); err != nil {
		return err
	}

	return nil
}
