package pipeline

import (
	"context"
	"time"

	"github.com/ashep/joex/internal/engine"
	"github.com/ashep/joex/internal/pipeline"
	proto "github.com/ashep/joex/sdk/proto/joex/v1"
	"github.com/rs/zerolog"
)

type pipelineService interface {
	Create(context.Context, pipeline.Pipeline) error
}

type Handler struct {
	svc pipelineService
	now func() time.Time
	l   zerolog.Logger
}

func New(svc pipelineService, now func() time.Time, l zerolog.Logger) *Handler {
	return &Handler{
		svc: svc,
		now: now,
		l:   l,
	}
}

func mapStepEngine(pt proto.Engine) engine.Type {
	switch pt {
	case proto.Engine_ENGINE_JS:
		return engine.JS
	default:
		return engine.Unknown
	}
}
