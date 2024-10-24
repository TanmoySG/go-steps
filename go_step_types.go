package gosteps

import (
	"encoding/json"
	"time"
)

// StepName type defined the name of the step
type StepName string
type BranchName string

// StepFn type defines the Step's Function
type StepFn func(ctx GoStepsCtx) StepResult
type ResolverFn func(ctx GoStepsCtx) BranchName

type RootStep struct {
	Steps Steps `json:"steps"`
}

// Step type defines a step with all configurations for the step
type Step struct {
	Name            StepName               `json:"name"`
	Function        StepFn                 `json:"-"`
	StepOpts        StepOpts               `json:"stepConfig"`
	Branches        *Branches              `json:"branches"`
	StepArgs        map[string]interface{} `json:"stepArgs"`
	StepResult      *StepResult            `json:"stepResult"`
	stepRunProgress stepRunProgress        `json:"-"`
}

type stepRunProgress struct {
	runCount int `json:"-"`
}

type Branch struct {
	BranchName BranchName `json:"branchName"`
	Steps      Steps      `json:"steps"`
}

type Steps []Step

type Branches struct {
	Branches []Branch   `json:"branches"`
	Resolver ResolverFn `json:"-"`
}

// step options
type StepOpts struct {
	ErrorsToRetry  []StepError   `json:"errorsToRetry"`
	RetryAllErrors bool          `json:"retryAllErrors"`
	MaxRunAttempts int           `json:"maxAttempts"`
	RetrySleep     time.Duration `json:"retrySleep"`
}

func (root *RootStep) ToJson() (string, error) {
	stepsBytes, err := json.Marshal(root)
	if err != nil {
		return "", err
	}

	return string(stepsBytes), nil
}
