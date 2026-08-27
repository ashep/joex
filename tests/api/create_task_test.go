//go:build functest

package api_test

import (
	"testing"

	jobproto "github.com/ashep/joex/sdk/proto/joex/v1"
	"github.com/ashep/joex/tests/testapp"
	"github.com/stretchr/testify/assert"
)

func TestCreateTask(main *testing.T) {
	main.Parallel()

	main.Run("InvalidAuthorization", func(t *testing.T) {
		t.Parallel()
		ta := testapp.New(t)

		cli := ta.TaskClient("anInvalidAuthToken")
		_, err := cli.CreateJob(t.Context(), &jobproto.CreateJobRequest{})

		assert.EqualError(t, err, "unauthenticated: unauthenticated")
		ta.AssertNoWarnsAndErrors()
	})
}
