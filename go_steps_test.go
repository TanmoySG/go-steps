package gosteps

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	stepError1 = StepError{
		StepErrorNameOrId: "error1",
		StepErrorMessage:  "error1",
	}
)

func Test_shouldRetry(t *testing.T) {

	testCases := []struct {
		StrictErrorCheck    bool
		Step                Step
		ExpectedShouldRetry bool
		ErrorToCheck        StepError
	}{
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
				},
				stepResult: &StepResult{
					StepState: StepStatePending,
				},
				stepRunProgress: StepRunProgress{
					runCount: 1,
				},
			},
			ExpectedShouldRetry: true,
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
				},
				stepRunProgress: StepRunProgress{
					runCount: 2,
				},
			},
			ExpectedShouldRetry: false,
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
					RetryAllErrors: true,
				},
				stepRunProgress: StepRunProgress{
					runCount: 1,
				},
				stepResult: &StepResult{
					StepState: StepStateFailed,
				},
			},
			ExpectedShouldRetry: false,
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
					RetryAllErrors: true,
				},
				stepResult: nil,
				stepRunProgress: StepRunProgress{
					runCount: 1,
				},
			},
			ExpectedShouldRetry: false,
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
					RetryAllErrors: true,
				},
				stepResult: &StepResult{
					StepState: StepStateError,
				},
				stepRunProgress: StepRunProgress{
					runCount: 1,
				},
			},
			ExpectedShouldRetry: true,
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
					RetryAllErrors: false,
					ErrorsToRetry: []StepError{
						stepError1,
					},
				},
				stepResult: &StepResult{
					StepState: StepStateError,
					StepError: &stepError1,
				},
				stepRunProgress: StepRunProgress{
					runCount: 1,
				},
			},
			ExpectedShouldRetry: true,
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
				},
				stepRunProgress: StepRunProgress{
					runCount: 1,
				},
				stepResult: &StepResult{
					StepState: StepStateComplete,
				},
			},
			ExpectedShouldRetry: false,
		},
	}

	for _, tc := range testCases {

		shouldRetry := tc.Step.shouldRetry()

		assert.Equal(t, tc.ExpectedShouldRetry, shouldRetry)
	}
}

func Test_shouldExit(t *testing.T) {

	testCases := []struct {
		Step               Step
		ExpectedShouldExit bool
	}{
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
				},
				stepResult: &StepResult{
					StepState: StepStateError,
				},
				stepRunProgress: StepRunProgress{
					runCount: 2,
				},
			},
			ExpectedShouldExit: true,
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
				},
				stepResult: &StepResult{
					StepState: StepStateComplete,
				},
				stepRunProgress: StepRunProgress{
					runCount: 2,
				},
			},
			ExpectedShouldExit: false,
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
				},
				stepResult: &StepResult{
					StepState: StepStateSkipped,
				},
				stepRunProgress: StepRunProgress{
					runCount: 2,
				},
			},
			ExpectedShouldExit: false,
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
				},
				stepResult: &StepResult{
					StepState: StepStatePending,
				},
				stepRunProgress: StepRunProgress{
					runCount: 2,
				},
			},
			ExpectedShouldExit: true,
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
				},
			},
		},
		{
			Step: Step{
				StepOpts: StepOpts{
					MaxRunAttempts: 2,
				},
				stepRunProgress: StepRunProgress{
					runCount: 1,
				},
				stepResult: &StepResult{
					StepState: StepStatePending,
				},
			},
		},
	}

	for _, tc := range testCases {

		shouldExit := tc.Step.shouldExit()

		assert.Equal(t, tc.ExpectedShouldExit, shouldExit)
	}
}

func Test_getExecutableBranch(t *testing.T) {

	b := Branches{
		Branches: []Branch{
			{
				BranchName: "branch1",
			},
			{
				BranchName: "branch2",
			},
		},
	}

	branch := b.getExecutableBranch("branch1")
	assert.Equal(t, "branch1", string(branch.BranchName))

	branch = b.getExecutableBranch("branch3")
	assert.Nil(t, branch)

}

func Test_setDefaults(t *testing.T) {

	step := Step{
		StepOpts: StepOpts{
			RetryAllErrors: true,
		},
	}

	step.setDefaults()
	assert.Equal(t, 1, step.StepOpts.MaxRunAttempts)
	assert.Nil(t, step.StepOpts.ErrorsToRetry)
}

func Test_setProgress(t *testing.T) {

	step := Step{}

	step.setProgress()
	assert.Equal(t, 1, step.stepRunProgress.runCount)
}

func Test_setResult(t *testing.T) {

	step := Step{}

	message := "step complete"
	sr := &StepResult{
		StepState:   StepStateComplete,
		StepMessage: &message,
		StepData: GoStepsCtxData{
			"key1": "value1",
		},
	}
	step.setResult(sr)

	assert.Equal(t, sr, step.stepResult)
}

func Test_Main(t *testing.T) {

	ctx := NewGoStepsContext()

	multipleDivide := Step{
		Name: "multipleDivide",
		Function: func(c GoStepsCtx) StepResult {
			res := c.GetData("result").(int) * 2
			return MarkStateComplete().WithData(map[string]interface{}{
				"result": res,
			})
		},
		Branches: &Branches{
			Resolver: func(ctx GoStepsCtx) BranchName {
				nx := ctx.GetData("result").(int)

				if nx%2 == 0 {
					return BranchName("divide")
				}
				return BranchName("multiple")
			},
			Branches: []Branch{
				{
					BranchName: "divide",
					Steps: Steps{
						{
							Name: "step3.divide",
							Function: func(c GoStepsCtx) StepResult {
								res := c.GetData("result").(int) / 2
								return MarkStateComplete().WithData(map[string]interface{}{
									"result": res,
								})
							},
						},
					},
				},
				{
					BranchName: "multiply",
					Steps: Steps{
						{
							Name: "step3.multiply",
							Function: func(c GoStepsCtx) StepResult {
								res := c.GetData("result").(int) * 2
								return MarkStateComplete().WithData(map[string]interface{}{
									"result": res,
								})
							},
						},
					},
				},
			},
		},
	}

	steps := Steps{
		{
			Name: "add",
			Function: func(c GoStepsCtx) StepResult {

				res := c.GetData("n1").(int) + c.GetData("n2").(int)
				return MarkStateComplete().WithData(map[string]interface{}{
					"result": res,
				})
			},
			StepArgs: map[string]interface{}{
				"n1": 5,
				"n2": 4,
			},
		},
		{
			Name: "subtract",
			Function: func(c GoStepsCtx) StepResult {
				res := c.GetData("n1").(int) - c.GetData("result").(int)
				return MarkStateComplete().WithData(map[string]interface{}{
					"result": res,
				}).WithMessage("step complete")
			},
			StepArgs: map[string]interface{}{
				"n1": 5,
			},
		},
		multipleDivide,
		{
			Name: "add",
			Function: func(c GoStepsCtx) StepResult {
				res := c.GetData("result").(int) + 5

				return MarkStateComplete().WithData(map[string]interface{}{
					"result": res,
				})
			},
		},
		{
			Name: "print",
			Function: func(c GoStepsCtx) StepResult {
				fmt.Println("result", c.GetData("result"))
				return MarkStateComplete()
			},
		},
	}

	root := NewStepChain(steps)

	root.Execute(ctx)

	ctxData := ctx.GetData("result")

	assert.Equal(t, 1, ctxData.(int))
}
