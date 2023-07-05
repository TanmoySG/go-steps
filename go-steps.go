package gosteps

import (
	"strings"
)

type StepName string

type Step struct {
	stepName         StepName
	function         interface{}
	additionalArgs   []interface{}
	nextSteps        []Step
	nextStepResolver interface{}
	errorsToRetry    []error
	strictErrorCheck bool
	skipRetry        bool
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
		// piping output from previous step and additional
		// arguments as arguments for current step
		stepArgs := []interface{}{}
		stepArgs = append(stepArgs, stepOutput...)
		stepArgs = append(stepArgs, step.additionalArgs...)

		// only runs for first step in steps
		if isEntryStep {
			stepArgs = initArgs
			isEntryStep = false
		}

		// execute current step passing step arguments
		stepOutput, stepError = step.function.(func(...interface{}) ([]interface{}, error))(stepArgs...)
		if stepError != nil {
			if !step.skipRetry && step.shouldRetry(stepError) {
				// piping args as output for re-running same step
				stepOutput = stepArgs
				continue
			}

			// skip retry and error not retryable
			return nil, stepError
		}

		// no next step, this is the final step
		if step.nextSteps == nil {
			finalOutput = stepOutput
			break
		}

		// next step is dependant on conditions
		if step.nextSteps != nil && step.nextStepResolver != nil {
			nextStepName := step.nextStepResolver.(func(...interface{}) StepName)(stepOutput...)
			step = *step.resolveNextStep(nextStepName)
			continue
		}

		// if there are multiple next steps but no resolver,
		// first one in the list is considered as next step
		step = step.nextSteps[0]
	}

	return finalOutput, nil
}

// should retry for error
func (step Step) shouldRetry(err error) bool {
	for _, errorToRetry := range step.errorsToRetry {
		if step.strictErrorCheck && err.Error() == errorToRetry.Error() {
			return true
		} else if !step.strictErrorCheck && strings.Contains(err.Error(), errorToRetry.Error()) {
			return true
		}
	}

	return false
}

// resolve next step by step name
func (step Step) resolveNextStep(stepName StepName) *Step {
	for _, nextStep := range step.nextSteps {
		if nextStep.stepName == stepName {
			return &nextStep
		}
	}

	return nil
}
