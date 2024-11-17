package gosteps

// StepState type defines the state
// of the step after execution
type StepState string

const (
	StepStateComplete StepState = "StepStateComplete" // step completed successfully             [non-retriable]
	StepStateFailed   StepState = "StepStateFailed"   // step failed to complete, without error  [non-retriable]
	StepStateSkipped  StepState = "StepStateSkipped"  // step was skipped                        [non-retriable]
	StepStatePending  StepState = "StepStatePending"  // step is pending, should be retried      [retriable]
	StepStateError    StepState = "StepStateError"    // step failed to complete, with error     [retriable]
)

// StepResult type defines the result of the step
type StepResult struct {
	StepData    GoStepsCtxData `json:"stepData"`            // stores the data from a step, if any
	StepState   StepState      `json:"stepState"`           // state of the step
	StepMessage *string        `json:"stepMessage"`         // message from the step execution, if any
	StepError   error          `json:"stepError,omitempty"` // error from the step execution, if any
}

// markState marks the state of the step
func markState(state StepState) StepResult {
	return StepResult{
		StepState: state,
	}
}

// MarkStateComplete marks the state of the step as complete
func MarkStateComplete() StepResult {
	return markState(StepStateComplete)
}

// MarkStateFailed marks the state of the step as failed
func MarkStateFailed() StepResult {
	return markState(StepStateFailed)
}

// MarkStateSkipped marks the state of the step as skipped
func MarkStateSkipped() StepResult {
	return markState(StepStateSkipped)
}

// MarkStatePending marks the state of the step as pending
func MarkStatePending() StepResult {
	return markState(StepStatePending)
}

// MarkStateError marks the state of the step as error
func MarkStateError() StepResult {
	return markState(StepStateError)
}

// WithData sets the data for the step
func (sr StepResult) WithData(data GoStepsCtxData) StepResult {
	sr.StepData = data
	return sr
}

// WithMessage sets the message for the step
func (sr StepResult) WithMessage(message string) StepResult {
	sr.StepMessage = &message
	return sr
}

// WithStepError sets the GoSteps - StepError for the step
func (sr StepResult) WithError(stepErr error) StepResult {
	sr.StepError = stepErr
	return sr
}
