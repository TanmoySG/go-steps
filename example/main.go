package main

import (
	"fmt"
	"os"
	"regexp"
	"time"

	gosteps "github.com/TanmoySG/go-steps"
	"github.com/rs/zerolog"
)

func main() {

	count := 0

	runLogFile, _ := os.OpenFile(
		"myapp.log",
		os.O_APPEND|os.O_CREATE|os.O_WRONLY,
		0664,
	)
	output := zerolog.MultiLevelWriter(os.Stdout, runLogFile)

	logger := gosteps.NewGoStepsLogger(output, &gosteps.LoggerOpts{StepLoggingEnabled: true})

	ctx := gosteps.NewGoStepsContext()

	ctx.Use(logger)

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
								c.Log("this is a message", gosteps.LogLevel(zerolog.ErrorLevel))
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
				c.Log("this is a message")

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
				if count < 2 {
					count++
					time.Sleep(2 * time.Second)
					return gosteps.MarkStatePending()
				}

				if count < 4 {
					count++
					return gosteps.MarkStateError().WithError(fmt.Errorf("errpr"))
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
				ErrorPatternsToRetry: []regexp.Regexp{
					*regexp.MustCompile("err*"),
				},
				ErrorsToRetry: []error{
					fmt.Errorf("errpr"),
				},
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
				c.Log(fmt.Sprintf("result %v", c.GetData("result")))
				return gosteps.MarkStateComplete()
			},
		},
	}

	stepsProcessor := gosteps.NewStepsProcessor(steps)
	stepsProcessor.Execute(ctx)
}
