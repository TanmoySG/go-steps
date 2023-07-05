package gosteps

import (
	"fmt"
	"strings"
)

var (
	unresolvedStepError = "error: step [%s] is unresolved, no step found with this name."
)

type StepName string

type Step struct {
	Name             StepName
	Function         interface{}
	AdditionalArgs   []interface{}
	NextSteps        []Step
	NextStepResolver interface{}
	ErrorsToRetry    []error
	StrictErrorCheck bool
	SkipRetry        bool
}

type Steps []Step

func (steps Steps) Execute(initArgs ...any) ([]interface{}, error) {
	// final output for steps execution
	var finalOutput []interface{}

	// initialize step output and step error
	var stepOutput []interface{}
	var stepError error

	// entry step
	var isEntryStep bool = true
	step := steps[0]

	for {
		// piping output from previous step as arguments for current step
		stepArgs := []interface{}{}
		stepArgs = append(stepArgs, stepOutput...)

		// only runs for first step in steps
		if isEntryStep {
			stepArgs = initArgs
			isEntryStep = false
		}

		// piping additional arguments  as arguments for current step
		stepArgs = append(stepArgs, step.AdditionalArgs...)

		// execute current step passing step arguments
		stepOutput, stepError = step.Function.(func(...interface{}) ([]interface{}, error))(stepArgs...)
		if stepError != nil {
			if !step.SkipRetry && step.shouldRetry(stepError) {
				// piping args as output for re-running same step
				stepOutput = stepArgs
				continue
			}

			// skip retry and error not retryable
			return nil, stepError
		}

		// no next step, this is the final step
		if step.NextSteps == nil {
			finalOutput = stepOutput
			return finalOutput, nil
		}

		// next step is dependant on conditions
		if step.NextSteps != nil && step.NextStepResolver != nil {
			nextStepName := step.NextStepResolver.(func(...interface{}) string)(stepOutput...)
			resolvedStep := step.resolveNextStep(StepName(nextStepName))
			if resolvedStep == nil {
				return stepOutput, fmt.Errorf(unresolvedStepError, step.Name)
			}

			step = *resolvedStep
			continue
		}

		// if there are multiple next steps but no resolver,
		// first one in the list is considered as next step
		step = step.NextSteps[0]
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
	for _, nextStep := range step.NextSteps {
		if nextStep.Name == stepName {
			return &nextStep
		}
	}

	return nil
}
