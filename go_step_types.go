package gosteps

import (
	"encoding/json"
	"time"
)

// StepName type defined the name of the step
type StepName string

// BranchName type defined the name of the branch
type BranchName string

// StepFn defines the Step's Function
type StepFn func(ctx GoStepsCtx) StepResult

// ResolverFn defines the Resolver Function
// to determine the branch to execute
type ResolverFn func(ctx GoStepsCtx) BranchName

// Step type defines a step with all configurations for the step
type Step struct {
	Name            StepName               `json:"name"`
	Function        StepFn                 `json:"-"`
	StepOpts        StepOpts               `json:"stepConfig"`
	Branches        *Branches              `json:"branches"`
	StepArgs        map[string]interface{} `json:"stepArgs"`
	StepResult      *StepResult            `json:"stepResult"` // make this private
	stepRunProgress stepRunProgress        `json:"-"`
}

// stepRunProgress type defines the progress of the step
// it contains the run/execution count of each step
type stepRunProgress struct {
	runCount int `json:"-"`
}

// Branch type defines a unique step-chain, of the step-tree
// Branches can be used to define different steps to be executed
// based on a resolver function
type Branch struct {
	BranchName BranchName `json:"branchName"`
	Steps      Steps      `json:"steps"`
}

// Steps type defines a list of steps
type Steps []Step

// Branches type defines a list of branches
// with a resolver function to determine the branch to execute
type Branches struct {
	Branches []Branch   `json:"branches"`
	Resolver ResolverFn `json:"-"`
}

// StepOpts type defines the configuration for the step
type StepOpts struct {
	ErrorsToRetry  []StepError   `json:"errorsToRetry"`
	RetryAllErrors bool          `json:"retryAllErrors"`
	MaxRunAttempts int           `json:"maxAttempts"`
	RetrySleep     time.Duration `json:"retrySleep"`
}

// ToJson converts the step-tree to JSON-string
func (branch *Branch) ToJson() (string, error) {
	stepsBytes, err := json.Marshal(branch)
	if err != nil {
		return "", err
	}

	return string(stepsBytes), nil
}

// NewStepChain creates a new root branch of the step-chain
func NewStepChain(steps Steps) *Branch {
	return &Branch{
		BranchName: "root",
		Steps:      steps,
	}
}
