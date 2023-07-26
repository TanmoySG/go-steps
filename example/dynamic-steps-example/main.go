package main

import (
	"fmt"

	gosteps "github.com/TanmoySG/go-steps"
	"github.com/TanmoySG/go-steps/example/funcs"
)

func main() {

	intsToAdd := []int{1, 4, 7, 10}

	var step *gosteps.Step
	for _, val := range intsToAdd {
		step = addStepToChain(step, funcs.Add, []interface{}{val})
	}

	finalOutput, err := step.Execute(1)
	if err != nil {
		fmt.Printf("error executing steps: %s, final output: [%s]\n", err, finalOutput)
	}

	fmt.Printf("Final Output: [%v]\n", finalOutput)
}

// step to add new next step to step-chain; basically a linked-list insertion
func addStepToChain(step *gosteps.Step, stepFunc gosteps.StepFn, additionalArgs []interface{}) *gosteps.Step {
	temp := gosteps.Step{
		Function:       stepFunc,
		StepArgs: additionalArgs,
	}

	if step == nil {
		step = &temp
		return step
	}

	curr := step
	for curr.NextStep != nil {
		curr = curr.NextStep
	}

	curr.NextStep = &temp
	return step
}
