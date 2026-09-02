package job

import (
	"context"
	"fmt"

	"github.com/ashep/joex/internal/apiutil"
	"github.com/ashep/joex/pkg/connecterr"
	proto "github.com/ashep/joex/sdk/proto/joex/v1"
)

func (h *Handler) CreateJob(ctx context.Context, req *proto.CreateJobRequest) (*proto.CreateJobResponse, error) {
	h.l.Info().Str("pipeline_id", req.GetPipelineId()).Msg("create job request")

	args, err := apiutil.MapProtoStructToDataType(req.GetArgs())
	if err != nil {
		return nil, connecterr.New(fmt.Errorf("validate args: %w", err), h.l)
	}

	job, err := h.jobSvc.CreateJob(ctx, req.GetPipelineId(), args)
	if err != nil {
		return nil, connecterr.New(err, h.l)
	}

	h.l.Info().Str("job_id", job.ID).Str("pipeline_id", req.GetPipelineId()).Msg("job created")

	return &proto.CreateJobResponse{
		JobId: job.ID,
	}, nil
}
