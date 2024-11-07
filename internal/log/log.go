package log

type StepLogStruct struct {
	Name     string
	State    string
	Message  *string
	Error    *string
	RunCount int
	MaxRun   int
	// Result   Result
	// Progress Progress
}

type Result struct {
	State   string
	Message *string
	Error   *string
}

type Progress struct {
	RunCount int
	MaxRun   int
}
