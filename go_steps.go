package gosteps

import (
	"fmt"
	"strings"
	"time"
)

func (step *Step) Execute(initArgs ...any) ([]interface{}, error) {
	// final output for step execution
	var finalOutput []interface{}

	// initialize step output and step error
	var stepOutput []interface{}
	var stepError error

	// no initial step or function
	if step == nil || step.Function == nil {
		return nil, nil
	}

	// entry step
	var isEntryStep bool = true

	// step reattepts
	var stepReAttemptsLeft int = step.MaxAttempts

	for {
		// piping output from previous step as arguments for current step
		var stepArgs []interface{}

		// only runs for first step in step
		if isEntryStep {
			step.StepArgs = append(step.StepArgs, initArgs...)
			isEntryStep = false
		}

		stepArgs = step.resolveStepArguments(stepOutput)

		// execute current step passing step arguments
		stepOutput, stepError = step.Function(stepArgs...)
		if stepError != nil {
			if !step.SkipRetry && step.shouldRetry(stepError) && stepReAttemptsLeft > 0 {
				// piping args as output for re-running same step
				stepOutput = stepArgs

				// decrementing re-attempts left for current run
				stepReAttemptsLeft -= 1

				// sleep step.RetrySleep duration if set
				if step.RetrySleep > 0 {
					time.Sleep(step.RetrySleep)
				}

				continue
			}

			// skip retry as step error not retryable
			// return output of previous step and error
			return stepArgs, stepError
		}

		// no next step, this is the final step
		if step.NextStep == nil && step.PossibleNextSteps == nil {
			finalOutput = stepOutput
			return finalOutput, nil
		}

		// next step is dependant on conditions
		if step.PossibleNextSteps != nil && step.NextStepResolver != nil {
			nextStepName := step.NextStepResolver.(func(...interface{}) string)(stepOutput...)
			resolvedStep := step.resolveNextStep(StepName(nextStepName))
			if resolvedStep == nil {
				return stepOutput, fmt.Errorf(unresolvedStepError, step.Name)
			}
			step.NextStep = resolvedStep
		}

		// set step as resolved or default nextStep
		step = step.NextStep

		// if step.MaxAttempts is not set, set default max value
		if step.MaxAttempts < 1 {
			step.MaxAttempts = DefaultMaxAttempts
		}

		// reset step re-attempts
		stepReAttemptsLeft = step.MaxAttempts - 1
	}
}

// should retry for error
func (step Step) shouldRetry(err error) bool {
	for _, errorToRetry := range step.ErrorsToRetry {
		if step.StrictErrorCheck && err.Error() == errorToRetry.Error() {
			return true
		} else if !step.StrictErrorCheck && strings.Contains(err.Error(), errorToRetry.Error()) {
			return true
		}
	}

	return false
}

// resolve next step by step name
func (step Step) resolveNextStep(stepName StepName) *Step {
	for _, nextStep := range step.PossibleNextSteps {
		if nextStep.Name == stepName {
			return &nextStep
		}
	}

	return nil
}

func (step Step) resolveStepArguments(previousStepReturns []interface{}) []interface{} {
	var resolvedStepArgs []interface{}

	switch step.UseArguments {
	case PreviousStepReturns:
		resolvedStepArgs = previousStepReturns
	case CurrentStepArgs:
		resolvedStepArgs = step.StepArgs
	case PreviousReturnsWithCurrentStepArgs:
		resolvedStepArgs = append(resolvedStepArgs, previousStepReturns...)
		resolvedStepArgs = append(resolvedStepArgs, step.StepArgs...)
	default: // covers UseCurrentStepArgsWithPreviousReturns too
		resolvedStepArgs = append(resolvedStepArgs, step.StepArgs...)
		resolvedStepArgs = append(resolvedStepArgs, previousStepReturns...)
	}

	return resolvedStepArgs
}
