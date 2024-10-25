package gosteps

// GoStepsCtxData type defines the data stored in the context
type GoStepsCtxData map[string]interface{}

// StepProgress type defines the progress of the step
type StepProgress struct {
	StepName   StepName   `json:"stepName"`
	StepResult StepResult `json:"stepResult"`
}

// GoStepsCtx type defines the context for the step-chain
type GoStepsCtx struct {
	data          GoStepsCtxData
	stepsProgress map[StepName]StepProgress
}

// GoStepsContext interface defines the methods for the context
type GoStepsContext interface {
	getCtx() GoStepsCtx
	SetData(key string, value interface{})
	GetData(key string) interface{}
	WithData(data map[string]interface{})
}

// GoStepsCtx type defines the context for the step-chain
func NewGoStepsContext() GoStepsContext {
	return GoStepsCtx{
		data:          GoStepsCtxData{},
		stepsProgress: map[StepName]StepProgress{},
	}
}

// getCtx returns the context - not exported
func (ctx GoStepsCtx) getCtx() GoStepsCtx {
	return ctx
}

// SetData sets the data in the context
func (ctx GoStepsCtx) SetData(key string, value interface{}) {
	ctx.data[key] = value
}

// GetData gets the data from the context
func (ctx GoStepsCtx) GetData(key string) interface{} {
	return ctx.data[key]
}

// WithData sets the data in the context
func (ctx GoStepsCtx) WithData(data map[string]interface{}) {
	for key, value := range data {
		ctx.SetData(key, value)
	}
}

// SetProgress sets the progress of the step
func (ctx GoStepsCtx) SetProgress(step StepName, stepResult StepResult) GoStepsCtx {
	ctx.stepsProgress[step] = StepProgress{
		StepName:   step,
		StepResult: stepResult,
	}

	return ctx
}

// GetProgress gets the progress of the step
func (ctx GoStepsCtx) GetProgress(step StepName) StepProgress {
	return ctx.stepsProgress[step]
}
