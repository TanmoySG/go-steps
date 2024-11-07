package gosteps

import (
	"fmt"
	"os"
	"text/template"

	l "github.com/TanmoySG/go-steps/internal/log"
)

const (
	stepResultErrorFormat = "[%s] %s"
)

func (step *Step) getStepLogStruct() l.StepLogStruct {
	var stepError string
	if step.stepResult.StepError != nil {
		stepError = fmt.Sprintf(
			stepResultErrorFormat,
			step.stepResult.StepError.StepErrorNameOrId,
			step.stepResult.StepError.StepErrorMessage,
		)
	}

	return l.StepLogStruct{
		Name: string(step.Name),

		State:   string(step.stepResult.StepState),
		Message: step.stepResult.StepMessage,
		Error:   &stepError,

		RunCount: step.stepRunProgress.runCount,
		MaxRun:   step.StepOpts.MaxRunAttempts,
	}

}

func (step *Step) log() {
	pattern := "Step: {{ .Name }} State: {{ .State }} {{ .RunCount }} \n"

	tmpl, err := template.New("log").Parse(pattern)
	if err != nil {
		fmt.Println("Error parsing template: ", err)
		return
	}

	err = tmpl.Execute(os.Stdout, step.getStepLogStruct())
	if err != nil {
		fmt.Println("Error executing template: ", err)
		return
	}
}
