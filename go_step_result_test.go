package gosteps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_markStatus(t *testing.T) {
	testCases := []struct {
		StepState StepState
		Function  func() StepResult
	}{

		{
			StepState: StepStateComplete,
			Function:  MarkStateComplete,
		},
		{
			StepState: StepStateFailed,
			Function:  MarkStateFailed,
		},
		{
			StepState: StepStateSkipped,
			Function:  MarkStateSkipped,
		},
		{
			StepState: StepStatePending,
			Function:  MarkStatePending,
		},
		{
			StepState: StepStateError,
			Function:  MarkStateError,
		},
	}

	for _, tc := range testCases {
		stepResult := tc.Function()
		assert.Equal(t, tc.StepState, stepResult.StepState)
	}
}

func Test_WithWrappedError_WithStepError(t *testing.T) {
	expectedWrappedError := &StepError{
		StepErrorNameOrId: "error",
		StepErrorMessage:  "error",
	}

	stepResult := StepResult{}.WithWrappedError(fmt.Errorf("error"))
	assert.Equal(t, expectedWrappedError, stepResult.StepError)

	stepResult = StepResult{}.WithStepError(stepError1)
	assert.Equal(t, &stepError1, stepResult.StepError)
}
