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

func Test_WithError(t *testing.T) {
	stepResult := StepResult{}.WithError(fmt.Errorf("error"))
	assert.Equal(t, fmt.Errorf("error"), stepResult.StepError)
}
