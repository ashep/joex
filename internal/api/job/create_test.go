package job_test

import (
	"context"
	"testing"
	"time"

	"github.com/ashep/go-app/testlogger"
	jobhandler "github.com/ashep/joex/internal/api/job"
	"github.com/ashep/joex/internal/datatype"
	"github.com/ashep/joex/internal/job"
	"github.com/ashep/joex/pkg/bufassert"
	"github.com/ashep/joex/pkg/connecterr"
	jobproto "github.com/ashep/joex/sdk/proto/joex/v1"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler_CreateJob(main *testing.T) {
	main.Parallel()

	setUp := func(t *testing.T) (*jobServiceMock, func() time.Time, *testlogger.Logger) {
		connecterr.ErrorCodeGenerator = func() string { return "123-321" }

		jobSvc := &jobServiceMock{}
		now := func() time.Time { return time.Unix(1234567890, 987654321) }
		l := testlogger.New(t)

		t.Cleanup(func() {
			jobSvc.AssertExpectations(t)
		})

		return jobSvc, now, l
	}

	main.Run("JobArgTypeUnexpected", func(t *testing.T) {
		t.Parallel()

		jobSvc, now, l := setUp(t)

		h := jobhandler.New(jobSvc, now, l.Logger())

		_, err := h.CreateJob(t.Context(), &jobproto.CreateJobRequest{
			Args: map[string]*jobproto.Arg{
				"foo": {
					Type: jobproto.DataType_DATA_TYPE_UNSPECIFIED,
				},
			},
		})

		bufassert.RequireInvalidArgumentError(t, err, "validate args: foo: unexpected type: unspecified")
		l.AssertNoWarnsAndErrors()
	})

	main.Run("JobArgTypeInvalid", func(t *testing.T) {
		t.Parallel()

		jobSvc, now, l := setUp(t)

		h := jobhandler.New(jobSvc, now, l.Logger())

		_, err := h.CreateJob(t.Context(), &jobproto.CreateJobRequest{
			Args: map[string]*jobproto.Arg{
				"foo": {
					Type:  jobproto.DataType_DATA_TYPE_INT,
					Value: "aNonInt",
				},
			},
		})

		bufassert.RequireInvalidArgumentError(t, err,
			`validate args: foo: aNonInt: cannot be converted to an int: strconv.Atoi: parsing "anonint": invalid syntax`)
		l.AssertNoWarnsAndErrors()
	})

	main.Run("JobServiceUnexpectedError", func(t *testing.T) {
		t.Parallel()

		jobSvc, now, l := setUp(t)
		jobSvc.On("CreateJob", mock.Anything, mock.Anything, mock.Anything).Return(job.Job{}, assert.AnError).Once()

		h := jobhandler.New(jobSvc, now, l.Logger())

		_, err := h.CreateJob(t.Context(), &jobproto.CreateJobRequest{})

		bufassert.RequireInternalError(t, err, "Error 123-321")
		l.AssertHasError(assert.AnError.Error())
	})

	main.Run("Success", func(t *testing.T) {
		t.Parallel()

		jobSvc, now, l := setUp(t)
		jobSvc.On("CreateJob", mock.Anything, mock.Anything, mock.Anything).Return(job.Job{
			ID: "aJobID",
		}, nil).Once()

		h := jobhandler.New(jobSvc, now, l.Logger())

		_, err := h.CreateJob(t.Context(), &jobproto.CreateJobRequest{
			PipelineId: "aPipelineID",
			Args: map[string]*jobproto.Arg{
				"foo": {
					Type:  jobproto.DataType_DATA_TYPE_INT,
					Value: "123",
				},
			},
		})

		require.NoError(t, err)
		l.AssertNoWarnsAndErrors()
		l.AssertContains(`{"level":"info","pipeline_id":"aPipelineID","message":"create job request"}
{"level":"info","job_id":"aJobID","pipeline_id":"aPipelineID","message":"job created"}`)
	})
}

type jobServiceMock struct {
	mock.Mock
}

func (m *jobServiceMock) CreateJob(
	ctx context.Context,
	pipeID string,
	args datatype.VarMap,
) (job.Job, error) {
	mArgs := m.Called(ctx, pipeID, args)
	return mArgs.Get(0).(job.Job), mArgs.Error(1)
}
