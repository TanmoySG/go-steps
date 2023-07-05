# GoSteps

GoSteps is a go library that helps in running functions as steps and reminds you to step out and get active (kidding!). The idea behind `gosteps` is to define set of functions as steps and execute them in a sequential fashion by piping output (other than error) from previous step, as arguments, into the next steps (not necessarily using the args).

## Usage

The `Step` type contains the requirments to execute a step function and move to next one.

```go
type Step struct {
 Name             StepName
 Function         interface{}
 AdditionalArgs   []interface{}
 NextSteps        []Step
 NextStepResolver interface{}
 ErrorsToRetry    []error
 StrictErrorCheck bool
 SkipRetry        bool
}
```

| Field            | Description                                                                                                  |
|------------------|--------------------------------------------------------------------------------------------------------------|
| Name             | Name of step                                                                                                 |
| Function         | The function to execute                                                                                      |
| AdditionalArgs   | any additional arguments need to pass to te step                                                             |
| NextSteps        | Candidate functions for next step (multiple next steps in-case of condition based execution)                 |
| NextStepResolver | A function that returns the step name, based on conditions, that is used to pick the nextStep from NextSteps |
| ErrorsToRetry    | A list of error to retry step for                                                                            |
| StrictErrorCheck | If set to `true` exact error is matched, else only presence of error is checked                              |
| SkipRetry        | If set to `true` step is not retried for any error                                                           |

## Defining Steps

To define steps, use the `gosteps.Steps` type and link the next steps in the `NextSteps` field as follows

```go
var steps = gosteps.Steps{
 {
  Name: "add",
  Function: funcs.Add,
  AdditionalArgs: []interface{}{2, 3},
  NextSteps: gosteps.Steps{
   {
    Name: "sub",
    Function:       funcs.Sub,
    AdditionalArgs: []interface{}{4},
   },
  },
 },
}
```

Here the first step is `Add` and next step (and final) is `Sub`, so the output of Add is piped to Sub and that gives the final output.

## Contditional Steps

Some steps might have multiple candidates for next step and the executable next step is to be picked based on the output of the current step. To do so, steps with multiple next step candidates must use the `NextStepResolver` field passing a resolver function that returns the Name of the function to use as next step.

The resolver function should be of type `func(args ...any) string`, where `args` are the output of current step and returned string is the name of the step to use.
```go
func nextStepResolver(args ...any) string {
 if args[0].(int) < 5 {
  fmt.Printf("StepResolver [%v]: Arguments is Negative, going with Multiply\n", args)
  return "add"
 }

 fmt.Printf("StepResolver [%v]: Arguments is Positive, going with Divide\n", args)
 return "sub"
}
```

## Executing Steps

To execute steps use the `Execute(initArgs ...any)` method, passing the (optional) initializing arguments.

```go
import (
 gosteps "github.com/TanmoySG/go-steps"
 "github.com/TanmoySG/go-steps/example/funcs"
)

func main() {
 initArgs := []interface{}{1, 2}
 finalOutput, err := steps.Execute(initArgs...)
 if err != nil {
  fmt.Printf("error executing steps: %s, final output: [%s]\n", err, finalOutput)
 }

 fmt.Printf("Final Output: [%v]\n", finalOutput)
}
```

## Retrying for Error

To retry a step for particular erors, use the `ErrorsToRetry` field passing the list of errors. To make sure the error matches exactly as that of the Errors to retry, pass `true` for the `StrictErrorCheck` field, otherwise only error-substring presense will be checked.

```go
ErrorsToRetry: []error{
    fmt.Errorf("error to retry"),
},
StrictErrorCheck: true
```

To skip retry on error pass `true` to the `SkipRetry` field.

**IMPORTANT:** There is no maximum retry for error, so if any error is encountered that is to be retried, it'll be retried till the error goes away and can lead to an infinite execution of the step. Please retry cautiously. Maximum retry parameter will be added soon.

## Example

In [this example](./example/main.go), we've used a set of complex steps with conditional step and retry. The flow of the same is

![](diag.png)

Execute the example steps

```
go run example/main.go

// output
Adding [1 2]
Sub [3 4]
StepResolver [[-1]]: Arguments is Negative, going with Multiply
Multiply [-1 -5]
Adding [5 100]
Running fake error function for arg [[105]]
Running fake error function for arg [[105]]
Running fake error function for arg [[105]]
Running fake error function for arg [[105]]
Multiply [3150 5250]
Final Output: [[16537500]]
```
