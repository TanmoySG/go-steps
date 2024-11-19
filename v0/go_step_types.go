package v0

import "time"

// StepName type defined the name of the step
type StepName string

// StepFn type defines the Step's Function
type StepFn func(...interface{}) ([]interface{}, error)

// PossibleNextSteps type is a list/array of Step objects
type PossibleNextSteps []Step

// Step type defines a step with all configurations for the step
type Step struct {
	Name              StepName
	Function          StepFn
	UseArguments      stepArgChainingType
	StepArgs          []interface{}
	NextStep          *Step
	PossibleNextSteps PossibleNextSteps
	NextStepResolver  interface{}
	ErrorsToRetry     []error
	StrictErrorCheck  bool
	SkipRetry         bool
	MaxAttempts       int
	RetrySleep        time.Duration
}

// enum type for step arguments chaining
type stepArgChainingType string
