package v0

import (
	"math"
)

const (
	// to avoid infinite runs due to the MaxAttempts not being set, we're keeping the default attempts to 100
	// if required, import and use the MaxMaxAttempts in the step.MaxAttempts field
	DefaultMaxAttempts = 100

	// the Max value is 9223372036854775807, which is not infinite but a huge number of attempts
	MaxMaxAttempts = math.MaxInt
)

var (
	// only previous step return will be passed to current step as arguments
	PreviousStepReturns stepArgChainingType = "PreviousStepReturns"

	// only current step arguments (StepArgs) will be passed to current step as arguments
	CurrentStepArgs stepArgChainingType = "CurrentStepArgs"

	// both previous step returns and current step arguments (StepArgs) will be passed
	// to current step as arguments - previous step returns, followed by current step args,
	PreviousReturnsWithCurrentStepArgs stepArgChainingType = "PreviousReturnsWithCurrentStepArgs"

	// both previous step returns and current step arguments (StepArgs) will be passed
	// to current step as arguments - current step args, followed by previous step returns
	CurrentStepArgsWithPreviousReturns stepArgChainingType = "CurrentStepArgsWithPreviousReturns"
)
