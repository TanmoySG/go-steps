package gosteps

// import (
// 	"fmt"
// 	"testing"

// 	"github.com/stretchr/testify/assert"
// )

// func Test_resolveStepArguments(t *testing.T) {

// 	samplePreviousStepOutput := []interface{}{4}
// 	sampleStepArgs := []interface{}{5}

// 	testCases := []struct {
// 		useArguments                  stepArgChainingType
// 		stepArgs                      []interface{}
// 		previousStepOutput            []interface{}
// 		expectedResolvedStepArguments []interface{}
// 	}{
// 		{
// 			useArguments:                  PreviousStepReturns,
// 			stepArgs:                      sampleStepArgs,
// 			previousStepOutput:            samplePreviousStepOutput,
// 			expectedResolvedStepArguments: samplePreviousStepOutput,
// 		},
// 		{
// 			useArguments:                  CurrentStepArgs,
// 			stepArgs:                      sampleStepArgs,
// 			previousStepOutput:            samplePreviousStepOutput,
// 			expectedResolvedStepArguments: sampleStepArgs,
// 		},
// 		{
// 			useArguments:                  PreviousReturnsWithCurrentStepArgs,
// 			stepArgs:                      sampleStepArgs,
// 			previousStepOutput:            samplePreviousStepOutput,
// 			expectedResolvedStepArguments: []interface{}{4, 5},
// 		},
// 		{
// 			useArguments:                  CurrentStepArgsWithPreviousReturns,
// 			stepArgs:                      sampleStepArgs,
// 			previousStepOutput:            samplePreviousStepOutput,
// 			expectedResolvedStepArguments: []interface{}{5, 4},
// 		},
// 	}

// 	for _, tc := range testCases {
// 		step := Step{
// 			StepArgs:     tc.stepArgs,
// 			UseArguments: tc.useArguments,
// 		}

// 		resolvedStepArgs := step.resolveStepArguments(tc.previousStepOutput)

// 		assert.Equal(t, tc.expectedResolvedStepArguments, resolvedStepArgs)
// 	}
// }

// func Test_resolveNextStep(t *testing.T) {
// 	step := Step{
// 		PossibleNextSteps: PossibleNextSteps{
// 			{
// 				Name: StepName("stepA"),
// 			},
// 			{
// 				Name: StepName("stepB"),
// 			},
// 			{
// 				Name: StepName("stepC"),
// 			},
// 		},
// 	}

// 	// happy path
// 	resolvedStep := step.resolveNextStep("stepA")
// 	assert.NotNil(t, resolvedStep)

// 	// step not found
// 	resolvedStep = step.resolveNextStep("stepD")
// 	assert.Nil(t, resolvedStep)
// }

// func Test_shouldRetry(t *testing.T) {

// 	testCases := []struct {
// 		StrictErrorCheck    bool
// 		ExpectedShouldRetry bool
// 		ErrorToCheck        error
// 	}{
// 		{
// 			StrictErrorCheck:    false,
// 			ErrorToCheck:        fmt.Errorf("error"),
// 			ExpectedShouldRetry: true,
// 		},
// 		{
// 			StrictErrorCheck:    true,
// 			ErrorToCheck:        fmt.Errorf("error"),
// 			ExpectedShouldRetry: false,
// 		},
// 		{
// 			StrictErrorCheck:    true,
// 			ErrorToCheck:        fmt.Errorf("error1"),
// 			ExpectedShouldRetry: true,
// 		},
// 		{
// 			StrictErrorCheck:    false,
// 			ErrorToCheck:        fmt.Errorf("wont retry for this error"),
// 			ExpectedShouldRetry: false,
// 		},
// 	}

// 	for _, tc := range testCases {
// 		step := Step{
// 			ErrorsToRetry: []error{
// 				fmt.Errorf("error1"),
// 				fmt.Errorf("error2"),
// 				fmt.Errorf("error3"),
// 			},
// 			StrictErrorCheck: tc.StrictErrorCheck,
// 		}

// 		shouldRetry := step.shouldRetry(tc.ErrorToCheck)

// 		assert.Equal(t, tc.ExpectedShouldRetry, shouldRetry)
// 	}
// }
