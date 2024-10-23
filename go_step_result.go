package gosteps

type StepState string

const (
	StepStateComplete StepState = "StepStateComplete" // step completed successfully
	StepStateFailed   StepState = "StepStateFailed"   // step failed to complete, without error
	StepStateSkipped  StepState = "StepStateSkipped"  // step was skipped
	StepStatePending  StepState = "StepStatePending"  // step is pending, should be retried
	StepStateError    StepState = "StepStateError"    // step failed to complete, with error
)

type StepResult struct {
	StepData    GoStepsCtxData `json:"stepData"`
	StepState   StepState      `json:"stepState"`
	StepMessage *string        `json:"stepMessage"`
	StepError   *StepError     `json:"stepError,omitempty"`
	runCount    int            `json:"-"`
}

type StepError struct {
	StepErrorNameOrId string `json:"stepErrorNameOrId"`
	StepErrorMessage  string `json:"stepErrorMessage"`
}

type StepProgress struct {
	StepName   StepName   `json:"stepName"`
	StepResult StepResult `json:"stepResult"`
}

func MarkState(state StepState) StepResult {
	return StepResult{
		StepState: state,
	}
}

func MarkStateComplete() StepResult {
	return MarkState(StepStateComplete)
}

func MarkStateFailed() StepResult {
	return MarkState(StepStateFailed)
}

func MarkStateSkipped() StepResult {
	return MarkState(StepStateSkipped)
}

func MarkStatePending() StepResult {
	return MarkState(StepStatePending)
}

func MarkStateError() StepResult {
	return MarkState(StepStateError)
}

func (sr StepResult) WithData(data GoStepsCtxData) StepResult {
	sr.StepData = data
	return sr
}

func (sr StepResult) WithMessage(message string) StepResult {
	sr.StepMessage = &message
	return sr
}

func (sr StepResult) WithWrappedError(e error) StepResult {
	sr.StepError = &StepError{
		StepErrorNameOrId: "error",
		StepErrorMessage:  e.Error(),
	}
	return sr
}

func (sr StepResult) WithStepError(stepErr StepError) StepResult {
	sr.StepError = &stepErr
	return sr
}
