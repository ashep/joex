package pipeline

import (
	"context"
	"fmt"

	"github.com/ashep/joex/internal/apiutil"
	"github.com/ashep/joex/internal/pipeline"
	"github.com/ashep/joex/internal/status"
	"github.com/ashep/joex/pkg/connecterr"
	"github.com/ashep/joex/pkg/typeutil"
	proto "github.com/ashep/joex/sdk/proto/joex/v1"
	"github.com/google/uuid"
)

func (h *Handler) CreatePipeline(
	ctx context.Context,
	req *proto.CreatePipelineRequest,
) (*proto.CreatePipelineResponse, error) {
	h.l.Info().Int("step_count", len(req.GetSteps())).Msg("create pipeline request")

	steps := make([]pipeline.Step, 0, len(req.GetSteps()))
	for i, reqStep := range req.GetSteps() {
		stepOpts, err := apiutil.MapProtoStructToDataType(reqStep.GetOpts())
		if err != nil {
			return nil, connecterr.New(fmt.Errorf("steps[%d].opts: %w", i, err), h.l)
		}

		step, err := pipeline.MakeStep(i, mapStepEngine(reqStep.GetEngine()), stepOpts, reqStep.GetAllowFail())
		if err != nil {
			return nil, connecterr.New(fmt.Errorf("steps[%d]: %w", i, err), h.l)
		}

		steps = append(steps, step)
	}

	now := h.now()
	p := pipeline.Pipeline{
		ID:        uuid.NewString(),
		Status:    status.Enabled,
		Steps:     steps,
		CreatedAt: typeutil.UnixTimeMs(now),
		UpdatedAt: typeutil.UnixTimeMs(now),
	}
	if err := p.Validate(); err != nil {
		return nil, connecterr.New(fmt.Errorf("new pipeline: %w", err), h.l)
	}

	if err := h.svc.Create(ctx, p); err != nil {
		h.l.Error().Err(err).Msg("create pipeline failed")
		return nil, connecterr.New(err, h.l)
	}

	h.l.Info().Str("pipeline_id", p.ID).Msg("pipeline created")

	return &proto.CreatePipelineResponse{
		Id: p.ID,
	}, nil
}
