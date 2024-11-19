package main

import (
	"fmt"
	"time"

	gosteps "github.com/TanmoySG/go-steps/v0"
	"github.com/TanmoySG/go-steps/example/v0/funcs"
)

const (
	stepMultiply = "Multiply"
	stepDivide   = "Divide"
)

// reading/maintaining this is a bit tricky will add
// a functional way to create this in the next version
var steps = gosteps.Step{
	Function: funcs.Add,
	StepArgs: []interface{}{2},
	NextStep: &gosteps.Step{
		Function:         funcs.Sub,
		StepArgs:         []interface{}{4},
		NextStepResolver: nextStepResolver,
		PossibleNextSteps: gosteps.PossibleNextSteps{
			{
				Name:     stepMultiply,
				Function: funcs.Multiply,
				StepArgs: []interface{}{-5},
				NextStep: &gosteps.Step{
					Function: funcs.Add,
					StepArgs: []interface{}{100},
					NextStep: &gosteps.Step{
						Function: funcs.StepWillError3Times,
						ErrorsToRetry: []error{
							fmt.Errorf("error"),
						},
						NextStep: &gosteps.Step{
							Function: funcs.StepWillErrorInfinitely,
							ErrorsToRetry: []error{
								fmt.Errorf("error"),
							},
							NextStep: &gosteps.Step{
								Function: funcs.Multiply,
							},
							StrictErrorCheck: false,
							MaxAttempts:      5, // use gosteps.MaxMaxAttempts for Maximum Possible reattempts
						},
						MaxAttempts: 5,
						RetrySleep:  1 * time.Second,
					},
				},
			},
			{
				Name:     stepDivide,
				Function: funcs.Divide,
				StepArgs: []interface{}{-2},
			},
		},
	},
}

func main() {
	initArgs := []interface{}{5}
	finalOutput, err := steps.Execute(initArgs...)
	if err != nil {
		fmt.Printf("error executing steps: %s, final output: [%s]\n", err, finalOutput)
	}

	fmt.Printf("Final Output: [%v]\n", finalOutput)
}

// step resolver
func nextStepResolver(args ...any) string {
	if args[0].(int) < 0 {
		fmt.Printf("StepResolver [%v]: Arguments is Negative, going with Multiply\n", args)
		return stepMultiply
	}

	fmt.Printf("StepResolver [%v]: Arguments is Positive, going with Divide\n", args)
	return stepDivide
}
