package main

import (
	"fmt"

	gosteps "github.com/TanmoySG/go-steps"
)

const (
	stepMultiply = "Multiply"
	stepDivide   = "Divide"
)

// reading/maintaining this is a bit tricky will add
// a functional way to create this in the next version
var steps = gosteps.Steps{}

func main() {

	intsToAdd := []int{1, 4, 7, 10}

	for _, i := range intsToAdd {
		fmt.Println(i)
	}

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
