package gosteps

import (
	"time"
)

// var Xchannel = make(chan GoStepsCtx)

func (rs *RootStep) Execute(c GoStepsContext) {
	rootBranch := Branch{
		BranchName: "root",
		Steps:      rs.Steps,
	}

	rootBranch.execute(c.getCtx())
}

func (branch *Branch) execute(c GoStepsCtx) {
	if branch.Steps == nil {
		return
	}

	branch.Steps.execute(c)
}

func (step *Step) setProgress() *Step {
	step.StepResult.runCount += 1
	return step
}

func (step *Step) setResult(stepResult *StepResult) *Step {
	step.StepResult = stepResult
	return step
}

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

	// Xchannel <- c
}

func (step *Step) setDefaults() {
	if step.StepOpts.MaxRunAttempts == 0 {
		step.StepOpts.MaxRunAttempts = 1
	}

	if step.StepOpts.RetryAllErrors {
		step.StepOpts.ErrorsToRetry = nil
	}
}

func (step *Step) sleep() {
	if step.StepOpts.RetrySleep > 0 {
		time.Sleep(step.StepOpts.RetrySleep)
	}
}

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
				branch.execute(c)
			}
		}

		currentStepCounter += 1
	}
}

func (branches *Branches) getExecutableBranch(branchName BranchName) *Branch {
	for _, branch := range branches.Branches {
		if branch.BranchName == branchName {
			return &branch
		}
	}

	return nil
}

// should retry for error
func (step *Step) shouldRetry() bool {
	if step.StepOpts.MaxRunAttempts == step.StepResult.runCount {
		return false
	}

	if step.StepResult.StepState == StepStatePending {
		return true
	}

	if step.StepOpts.RetryAllErrors {
		return true
	}

	if step.StepResult.StepState == StepStateError {
		for _, errorToRetry := range step.StepOpts.ErrorsToRetry {
			if errorToRetry == *step.StepResult.StepError {
				return true
			}
		}
	}

	return false
}

func (step *Step) shouldExit() bool {
	if step.StepOpts.MaxRunAttempts == step.StepResult.runCount {
		switch step.StepResult.StepState {
		case StepStateComplete, StepStateSkipped:
			return false
		default: // StepStateError, StepStatePending, StepStateFailed
			return true
		}
	}

	return false
}
