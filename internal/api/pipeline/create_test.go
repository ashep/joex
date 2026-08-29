package pipeline_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/ashep/go-app/testlogger"
	pipelinehandler "github.com/ashep/joex/internal/api/pipeline"
	"github.com/ashep/joex/internal/engine"
	"github.com/ashep/joex/internal/pipeline"
	"github.com/ashep/joex/internal/status"
	"github.com/ashep/joex/pkg/bufassert"
	"github.com/ashep/joex/pkg/connecterr"
	"github.com/ashep/joex/pkg/typeutil"
	pipelineproto "github.com/ashep/joex/sdk/proto/joex/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_CreatePipeline(main *testing.T) {
	main.Parallel()

	setUp := func(t *testing.T) (*pipelineServiceMock, func() time.Time, *testlogger.Logger) {
		connecterr.ErrorCodeGenerator = func() string { return "123-321" }

		pipelineSvc := &pipelineServiceMock{}
		now := func() time.Time { return time.Unix(1234567890, 987654321) }
		l := testlogger.New(t)

		t.Cleanup(func() {
			pipelineSvc.AssertExpectations(t)
		})

		return pipelineSvc, now, l
	}

	// validStep is the minimal step a pipeline needs to pass validation.
	validStep := func() *pipelineproto.Step {
		return &pipelineproto.Step{
			Engine: pipelineproto.Engine_ENGINE_JS,
			Opts: map[string]*pipelineproto.Arg{
				"code": {
					Type:  pipelineproto.DataType_DATA_TYPE_STRING,
					Value: "1+1",
				},
			},
		}
	}

	main.Run("StepOptsArgTypeUnexpected", func(t *testing.T) {
		t.Parallel()

		pipelineSvc, now, l := setUp(t)

		h := pipelinehandler.New(pipelineSvc, now, l.Logger())

		_, err := h.CreatePipeline(t.Context(), &pipelineproto.CreatePipelineRequest{
			Steps: []*pipelineproto.Step{
				{
					Opts: map[string]*pipelineproto.Arg{
						"foo": {
							Type: pipelineproto.DataType_DATA_TYPE_UNSPECIFIED,
						},
					},
				},
			},
		})

		bufassert.RequireInvalidArgumentError(t, err, "steps[0].opts: foo: unexpected type: unspecified")
		l.AssertNoWarnsAndErrors()
	})

	main.Run("StepOptsArgTypeInvalid", func(t *testing.T) {
		t.Parallel()

		pipelineSvc, now, l := setUp(t)

		h := pipelinehandler.New(pipelineSvc, now, l.Logger())

		_, err := h.CreatePipeline(t.Context(), &pipelineproto.CreatePipelineRequest{
			Steps: []*pipelineproto.Step{
				{
					Opts: map[string]*pipelineproto.Arg{
						"foo": {
							Type:  pipelineproto.DataType_DATA_TYPE_INT,
							Value: "aNonInt",
						},
					},
				},
			},
		})

		bufassert.RequireInvalidArgumentError(t, err,
			`steps[0].opts: foo: aNonInt: cannot be converted to an int: strconv.Atoi: parsing "anonint": invalid syntax`)
		l.AssertNoWarnsAndErrors()
	})

	main.Run("StepEngineUnsupported", func(t *testing.T) {
		t.Parallel()

		pipelineSvc, now, l := setUp(t)

		h := pipelinehandler.New(pipelineSvc, now, l.Logger())

		_, err := h.CreatePipeline(t.Context(), &pipelineproto.CreatePipelineRequest{
			Steps: []*pipelineproto.Step{
				{
					Engine: pipelineproto.Engine_ENGINE_UNSPECIFIED,
				},
			},
		})

		bufassert.RequireInvalidArgumentError(t, err, "steps[0]: unspecified: unsupported engine")
		l.AssertNoWarnsAndErrors()
	})

	main.Run("NoSteps", func(t *testing.T) {
		t.Parallel()

		pipelineSvc, now, l := setUp(t)

		h := pipelinehandler.New(pipelineSvc, now, l.Logger())

		_, err := h.CreatePipeline(t.Context(), &pipelineproto.CreatePipelineRequest{})

		bufassert.RequireInvalidArgumentError(t, err, "new pipeline: steps: must have at least one step")
		l.AssertNoWarnsAndErrors()
	})

	main.Run("PipelineServiceUnexpectedError", func(t *testing.T) {
		t.Parallel()

		pipelineSvc, now, l := setUp(t)
		pipelineSvc.On("Create", mock.Anything, mock.Anything).Return(assert.AnError).Once()

		h := pipelinehandler.New(pipelineSvc, now, l.Logger())

		_, err := h.CreatePipeline(t.Context(), &pipelineproto.CreatePipelineRequest{
			Steps: []*pipelineproto.Step{validStep()},
		})

		bufassert.RequireInternalError(t, err, "Error 123-321")
		l.AssertHasError(assert.AnError.Error())
	})

	main.Run("Success", func(t *testing.T) {
		t.Parallel()

		pipelineSvc, now, l := setUp(t)

		var created pipeline.Pipeline
		pipelineSvc.On("Create", mock.Anything, mock.Anything).
			Run(func(args mock.Arguments) {
				created = args.Get(1).(pipeline.Pipeline)
			}).
			Return(nil).Once()

		h := pipelinehandler.New(pipelineSvc, now, l.Logger())

		step := validStep()
		step.AllowFail = true

		resp, err := h.CreatePipeline(t.Context(), &pipelineproto.CreatePipelineRequest{
			Steps: []*pipelineproto.Step{step},
		})

		require.NoError(t, err)
		require.NotEmpty(t, resp.GetId())
		assert.Equal(t, created.ID, resp.GetId())
		assert.Equal(t, status.EnabledStatus(status.Enabled), created.Status)
		assert.Equal(t, typeutil.UnixTimeMs(now()), created.CreatedAt)
		assert.Equal(t, typeutil.UnixTimeMs(now()), created.UpdatedAt)

		require.Len(t, created.Steps, 1)
		assert.Equal(t, 0, created.Steps[0].ID)
		assert.Equal(t, engine.JS, created.Steps[0].Engine)
		assert.True(t, created.Steps[0].AllowFail)

		l.AssertNoWarnsAndErrors()
		l.AssertContains(fmt.Sprintf(`{"level":"info","step_count":1,"message":"create pipeline request"}
{"level":"info","pipeline_id":"%s","message":"pipeline created"}`, created.ID))
	})
}

type pipelineServiceMock struct {
	mock.Mock
}

func (m *pipelineServiceMock) Create(ctx context.Context, p pipeline.Pipeline) error {
	args := m.Called(ctx, p)
	return args.Error(0)
}
