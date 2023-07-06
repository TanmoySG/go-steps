package main

import (
	"fmt"
	"time"

	gosteps "github.com/TanmoySG/go-steps"
	"github.com/TanmoySG/go-steps/example/funcs"
)

const (
	stepMultiply = "Multiply"
	stepDivide   = "Divide"
)

// reading/maintaining this is a bit tricky will add
// a functional way to create this in the next version
var steps = gosteps.Steps{
	{
		Function: funcs.Add,
		NextSteps: gosteps.Steps{
			{
				Function:         funcs.Sub,
				AdditionalArgs:   []interface{}{4},
				NextStepResolver: nextStepResolver,
				NextSteps: gosteps.Steps{
					{
						Name:           stepMultiply,
						Function:       funcs.Multiply,
						AdditionalArgs: []interface{}{-5},
						NextSteps: gosteps.Steps{
							{
								Function:       funcs.Add,
								AdditionalArgs: []interface{}{100},
								NextSteps: gosteps.Steps{
									{
										Function: funcs.StepWillError3Times,
										ErrorsToRetry: []error{
											fmt.Errorf("error"),
										},
										NextSteps: gosteps.Steps{
											{
												Function: funcs.StepWillErrorInfinitely,
												ErrorsToRetry: []error{
													fmt.Errorf("error"),
												},
												NextSteps: gosteps.Steps{
													{
														Function: funcs.Multiply,
													},
												},
												StrictErrorCheck: true,
												MaxAttempts:      5, // use gosteps.MaxMaxAttempts for Maximum Possible reattempts
											},
										},
										MaxAttempts: 5,
										RetrySleep:  1 * time.Second,
									},
								},
							},
						},
					},
					{
						Name:           stepDivide,
						Function:       funcs.Divide,
						AdditionalArgs: []interface{}{-2},
					},
				},
			},
		},
	},
}

func main() {
	initArgs := []interface{}{1, 2}
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
