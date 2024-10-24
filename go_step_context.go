package gosteps

import (
	"encoding/json"
)

type GoStepsCtxData map[string]interface{}

type GoStepsCtx struct {
	Data          GoStepsCtxData            `json:"data"`
	StepsProgress map[StepName]StepProgress `json:"stepsProgress"`
}

type GoStepsContext interface {
	getCtx() GoStepsCtx
	SetData(key string, value interface{})
	GetData(key string) interface{}
	WithData(data map[string]interface{})
}

func NewGoStepsContext() GoStepsContext {
	return GoStepsCtx{
		Data:          GoStepsCtxData{},
		StepsProgress: map[StepName]StepProgress{},
	}
}

func (ctx GoStepsCtx) ToJson() (string, error) {
	cBytes, err := json.Marshal(ctx)
	if err != nil {
		return "", err
	}

	return string(cBytes), nil
}

func (ctx GoStepsCtx) getCtx() GoStepsCtx {
	return ctx
}

func (ctx GoStepsCtx) SetData(key string, value interface{}) {
	ctx.Data[key] = value
}

func (ctx GoStepsCtx) GetData(key string) interface{} {
	return ctx.Data[key]
}

func (ctx GoStepsCtx) WithData(data map[string]interface{}) {
	for key, value := range data {
		ctx.SetData(key, value)
	}
}

func (ctx GoStepsCtx) SetProgress(step StepName, stepResult StepResult) GoStepsCtx {
	ctx.StepsProgress[step] = StepProgress{
		StepName:   step,
		StepResult: stepResult,
	}

	return ctx
}

func (ctx GoStepsCtx) GetProgress(step StepName) StepProgress {
	return ctx.StepsProgress[step]
}

// func (ctx GoStepsCtx) SetStepRetires(stepName StepName) GoStepsCtx {
// 	progress := ctx.StepsProgress[stepName]
// 	progress.StepRetries += 1
// 	ctx.StepsProgress[stepName] = progress
// 	return ctx
// }
