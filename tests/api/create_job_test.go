//go:build functest

package api_test

import (
	"testing"

	jobproto "github.com/ashep/joex/sdk/proto/joex/v1"
	"github.com/ashep/joex/tests/testapp"
	"github.com/stretchr/testify/assert"
)

func TestCreateJob(main *testing.T) {
	main.Parallel()

	main.Run("EmptyAuthorizationToken", func(t *testing.T) {
		t.Parallel()
		ta := testapp.New(t)

		cli := ta.NewJobServiceClient("")
		_, err := cli.CreateJob(t.Context(), &jobproto.CreateJobRequest{})

		assert.EqualError(t, err, "unauthenticated: unauthenticated")
		ta.AssertNoWarnsAndErrors()
	})

	main.Run("InvalidAuthorizationToken", func(t *testing.T) {
		t.Parallel()
		ta := testapp.New(t)

		cli := ta.NewJobServiceClient("anInvalidAuthToken")
		_, err := cli.CreateJob(t.Context(), &jobproto.CreateJobRequest{})

		assert.EqualError(t, err, "unauthenticated: unauthenticated")
		ta.AssertNoWarnsAndErrors()
	})
}
