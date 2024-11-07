package main

import (
	"fmt"

	gosteps "github.com/TanmoySG/go-steps"
)

func main() {

	count := 0

	ctx := gosteps.NewGoStepsContext()

	multipleDivide := gosteps.Step{
		Name: "multipleDivide",
		Function: func(c gosteps.GoStepsCtx) gosteps.StepResult {
			res := c.GetData("result").(int) * 2
			return gosteps.MarkStateComplete().WithData(map[string]interface{}{
				"result": res,
			})
		},
		Branches: &gosteps.Branches{
			Resolver: func(ctx gosteps.GoStepsCtx) gosteps.BranchName {
				nx := ctx.GetData("result").(int)

				if nx%2 == 0 {
					return gosteps.BranchName("divide")
				}
				return gosteps.BranchName("multiple")
			},
			Branches: []gosteps.Branch{
				{
					BranchName: "divide",
					Steps: gosteps.Steps{
						{
							Name: "step3.divide",
							Function: func(c gosteps.GoStepsCtx) gosteps.StepResult {
								res := c.GetData("result").(int) / 2
								return gosteps.MarkStateComplete().WithData(map[string]interface{}{
									"result": res,
								})
							},
						},
					},
				},
				{
					BranchName: "multiply",
					Steps: gosteps.Steps{
						{
							Name: "step3.multiply",
							Function: func(c gosteps.GoStepsCtx) gosteps.StepResult {
								res := c.GetData("result").(int) * 2
								return gosteps.MarkStateComplete().WithData(map[string]interface{}{
									"result": res,
								})
							},
						},
					},
				},
			},
		},
	}

	steps := gosteps.Steps{
		{
			Name: "add",
			Function: func(c gosteps.GoStepsCtx) gosteps.StepResult {

				res := c.GetData("n1").(int) + c.GetData("n2").(int)
				return gosteps.MarkStateComplete().WithData(map[string]interface{}{
					"result": res,
				})
			},
			StepArgs: map[string]interface{}{
				"n1": 5,
				"n2": 4,
			},
		},
		{
			Name: "subtract",
			Function: func(c gosteps.GoStepsCtx) gosteps.StepResult {
				if count < 3 {
					count++
					return gosteps.MarkStateError().WithWrappedError(fmt.Errorf("error"))
				}
				res := c.GetData("n1").(int) - c.GetData("result").(int)
				return gosteps.MarkStateComplete().WithData(map[string]interface{}{
					"result": res,
				})
			},
			StepArgs: map[string]interface{}{
				"n1": 5,
			},
			StepOpts: gosteps.StepOpts{
				MaxRunAttempts: 5,
				RetryAllErrors: true,
			},
		},
		multipleDivide,
		{
			Name: "add",
			Function: func(c gosteps.GoStepsCtx) gosteps.StepResult {
				res := c.GetData("result").(int) + 5

				return gosteps.MarkStateComplete().WithData(map[string]interface{}{
					"result": res,
				})
			},
		},
		{
			Name: "print",
			Function: func(c gosteps.GoStepsCtx) gosteps.StepResult {
				fmt.Println("result", c.GetData("result"))
				return gosteps.MarkStateComplete()
			},
		},
	}

	stepsProcessor := gosteps.NewStepsProcessor(steps)
	stepsProcessor.Execute(ctx)
}
