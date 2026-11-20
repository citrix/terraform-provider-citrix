// Copyright © 2026. Citrix Systems, Inc.

package util

import (
	"testing"

	"github.com/citrix/citrix-daas-rest-go/citrixquickcreate"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newDeploymentTask(state citrixquickcreate.TaskState) *citrixquickcreate.DeploymentTask {
	task := citrixquickcreate.NewDeploymentTask(citrixquickcreate.TASKTYPE_DEPLOYMENT)
	task.SetTaskState(state)
	return task
}

func TestFailedQcsTask(t *testing.T) {
	failedConnection := citrixquickcreate.NewResourceConnectionTask(citrixquickcreate.TASKTYPE_RESOURCE_CONNECTION)
	failedConnection.SetTaskState(citrixquickcreate.TASKSTATE_ERROR)
	failedDeployment := newDeploymentTask(citrixquickcreate.TASKSTATE_ERROR)

	tests := []struct {
		name     string
		response *citrixquickcreate.GetTaskAsync200Response
		expected *citrixquickcreate.TaskBase
	}{
		{name: "nil response", response: nil, expected: nil},
		{name: "failed resource connection task", response: &citrixquickcreate.GetTaskAsync200Response{ResourceConnectionTask: failedConnection}, expected: &failedConnection.TaskBase},
		{name: "failed deployment task", response: &citrixquickcreate.GetTaskAsync200Response{DeploymentTask: failedDeployment}, expected: &failedDeployment.TaskBase},
		{name: "running deployment task", response: &citrixquickcreate.GetTaskAsync200Response{DeploymentTask: newDeploymentTask(citrixquickcreate.TASKSTATE_PROCESSING)}, expected: nil},
		{name: "completed deployment task", response: &citrixquickcreate.GetTaskAsync200Response{DeploymentTask: newDeploymentTask(citrixquickcreate.TASKSTATE_COMPLETED)}, expected: nil},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			actual := failedQcsTask(tc.response)
			if tc.expected == nil {
				assert.Nil(t, actual)
				return
			}
			assert.Same(t, tc.expected, actual)
		})
	}
}

func TestQcsTaskError(t *testing.T) {
	withDetails := newDeploymentTask(citrixquickcreate.TASKSTATE_ERROR)
	withDetails.SetStatus("Deployment failed")
	credentialsError := citrixquickcreate.TaskError{}
	credentialsError.SetMessage("The domain credentials were rejected")
	withDetails.SetErrors([]citrixquickcreate.TaskError{credentialsError, {}})

	tests := []struct {
		name     string
		task     *citrixquickcreate.TaskBase
		expected string
	}{
		{
			name:     "status and error messages",
			task:     &withDetails.TaskBase,
			expected: "task task-1 ended in state ERROR: Deployment failed; The domain credentials were rejected",
		},
		{
			name:     "no details",
			task:     &newDeploymentTask(citrixquickcreate.TASKSTATE_ERROR).TaskBase,
			expected: "task task-1 ended in state ERROR",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := qcsTaskError("task-1", tc.task)
			require.Error(t, err)
			assert.Equal(t, tc.expected, err.Error())
		})
	}
}
