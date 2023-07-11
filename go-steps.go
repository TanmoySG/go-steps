package gosteps

import (
	"fmt"
	"math"
	"strings"
	"time"
)

const (
	// to avoid infinite runs due to the MaxAttempts not being set, we're keeping the default attempts to 100
	// if required, import and use the MaxMaxAttempts in the step.MaxAttempts field
	DefaultMaxAttempts = 100

	// the Max value is 9223372036854775807, which is not infinite but a huge number of attempts
	MaxMaxAttempts = math.MaxInt
)

var (
	unresolvedStepError = "error: step [%s] is unresolved, no step found with this name."
)

type StepName string

type Step struct {
	Name              StepName
	Function          interface{}
	AdditionalArgs    []interface{}
	NextStep          *Step
	PossibleNextSteps []Step
	NextStepResolver  interface{}
	ErrorsToRetry     []error
	StrictErrorCheck  bool
	SkipRetry         bool
	MaxAttempts       int
	RetrySleep        time.Duration
}

type Steps []Step

func (steps *Step) Execute(initArgs ...any) ([]interface{}, error) {
	// final output for steps execution
	var finalOutput []interface{}

	// initialize step output and step error
	var stepOutput []interface{}
	var stepError error

	// no entry or initial step
	if steps == nil {
		return nil, nil
	}

	// entry step
	var isEntryStep bool = true
	step := steps

	// step reattepts
	var stepReAttemptsLeft int = step.MaxAttempts

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
		if step.NextStep == nil || step.PossibleNextSteps == nil {
			finalOutput = stepOutput
			return finalOutput, nil
		}

		// if there are multiple next steps but no resolver,
		// first one in the list is considered as next step

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
