package job

import (
	"context"

	"github.com/ashep/joex/internal/apiutil"
	"github.com/ashep/joex/pkg/connecterr"
	proto "github.com/ashep/joex/sdk/proto/joex/v1"
)

func (h *Handler) CreateJob(ctx context.Context, req *proto.CreateJobRequest) (*proto.CreateJobResponse, error) {
	h.l.Info().Str("pipeline_id", req.GetPipelineId()).Msg("create job request")

	args, err := apiutil.MapProtoDataTypeArgs(req.GetArgs())
	if err != nil {
		return nil, connecterr.New(err, h.now)
	}

	job, err := h.jobSvc.CreateJob(ctx, req.GetPipelineId(), args)
	if err != nil {
		h.l.Error().Err(err).Str("pipeline_id", req.GetPipelineId()).Msg("create job failed")
		return nil, connecterr.New(err, h.now)
	}

	h.l.Info().Str("job_id", job.ID).Str("pipeline_id", req.GetPipelineId()).Msg("job created")

	return &proto.CreateJobResponse{
		JobId: job.ID,
	}, nil
}
