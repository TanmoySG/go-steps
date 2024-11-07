package gosteps

import (
	"time"
)

// Execute a branch with the context provided
func (branch *Branch) Execute(c GoStepsContext) {
	if branch.Steps == nil {
		return
	}

	branch.Steps.execute(c.getCtx())
}

// setProgress sets the run progress (runCount) of a step
func (step *Step) setProgress() {
	step.stepRunProgress.runCount += 1
}

// setResult sets the result of the executed step
func (step *Step) setResult(stepResult *StepResult) *Step {
	step.stepResult = stepResult
	return step
}

// setDefaults sets the default values for the StepOpts
func (step *Step) setDefaults() {
	if step.StepOpts.MaxRunAttempts == 0 {
		step.StepOpts.MaxRunAttempts = 1
	}

	if step.StepOpts.RetryAllErrors {
		step.StepOpts.ErrorsToRetry = nil
	}
}

// sleep for the retry sleep duration of the step
func (step *Step) sleep() {
	if step.StepOpts.RetrySleep > 0 {
		time.Sleep(step.StepOpts.RetrySleep)
	}
}

// Execute a step with the context provided
func (step *Step) execute(c GoStepsCtx) {
	if step.Function == nil {
		return
	}

	c.WithData(step.StepArgs)
	step.setDefaults()

	stepResult := step.Function(c)
	step.setResult(&stepResult)

	c.WithData(stepResult.StepData)
	c.SetProgress(step.Name, stepResult)

	step.setProgress()

	step.log()
}

// Execute a chain of steps with the context provided
func (steps *Steps) execute(c GoStepsCtx) {
	s := *steps
	if len(s) == 0 {
		return
	}

	currentStepCounter := 0

	var currentStep *Step = &s[currentStepCounter]
	for currentStep != nil {
		if currentStepCounter >= len(s) {
			break
		}

		currentStep = &s[currentStepCounter]
		currentStep.execute(c)

		if currentStep.shouldRetry() {
			currentStep.sleep()
			continue
		}

		if currentStep.shouldExit() {
			break
		}

		branches := currentStep.Branches
		if branches != nil {
			branchName := branches.Resolver(c)
			branch := branches.getExecutableBranch(branchName)

			if branch != nil {
				branch.Execute(c)
			}
		}

		currentStepCounter += 1
	}
}

// getExecutableBranch returns the branch to execute based on the resolver result
func (branches *Branches) getExecutableBranch(branchName BranchName) *Branch {
	for _, branch := range branches.Branches {
		if branch.BranchName == branchName {
			return &branch
		}
	}

	return nil
}

// shouldRetry checks if the step should be retried
// retry steps, if:
//   - step state is pending
//   - step state is error and RetryAllErrors is true
//   - step state is error and error is in ErrorsToRetry
//   - step run count is less than MaxRunAttempts
//
// skip retry if:
//   - step state is failed, complete or skipped
//   - step run count is equal to MaxRunAttempts
func (step *Step) shouldRetry() bool {
	if step.StepOpts.MaxRunAttempts == step.stepRunProgress.runCount {
		return false
	}

	if step.stepResult == nil {
		return false
	}

	if step.stepResult.StepState == StepStateFailed {
		return false
	}

	if step.stepResult.StepState == StepStatePending {
		return true
	}

	if step.stepResult.StepState == StepStateError && step.StepOpts.RetryAllErrors {
		return true
	}

	if step.stepResult.StepState == StepStateError {
		for _, errorToRetry := range step.StepOpts.ErrorsToRetry {
			if errorToRetry == *step.stepResult.StepError {
				return true
			}
		}
	}

	return false
}

// shouldExit checks if the step should exists
// and step-chain execution should be stopped
func (step *Step) shouldExit() bool {
	if step.stepResult == nil {
		return false
	}

	switch step.stepResult.StepState {
	case StepStateComplete, StepStateSkipped:
		return false
	default: // StepStateError, StepStatePending, StepStateFailed
		return true
	}
}
