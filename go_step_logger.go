package gosteps

import (
	"io"
	"os"

	// "strings"

	"github.com/rs/zerolog"
)

type LogLevel zerolog.Level

var (
	// defaultLogOut is the default output for the logger
	defaultLogOut = os.Stdout

	// stateToLevelMap maps the StepState to the log level
	stateToLevelMap = map[StepState]zerolog.Level{
		StepStateComplete: zerolog.InfoLevel,
		StepStateFailed:   zerolog.WarnLevel,
		StepStateSkipped:  zerolog.DebugLevel,
		StepStatePending:  zerolog.DebugLevel,
		StepStateError:    zerolog.ErrorLevel,
	}

	// Log Level of GoSteps Logger implementation of zerolog.Level
	DebugLevel = LogLevel(zerolog.DebugLevel)
	InfoLevel  = LogLevel(zerolog.InfoLevel)
	WarnLevel  = LogLevel(zerolog.WarnLevel)
	ErrorLevel = LogLevel(zerolog.ErrorLevel)
)

type goStepsLogger struct {
	config *LoggerOpts
	logger zerolog.Logger
}

type LoggerOpts struct {
	StepLoggingEnabled bool
}

// NewGoStepsLogger returns a new instance of the GoStepsLogger
//
// output: is of type io.Writer, example os.Stdout, for more options refer
// to zerolog documentation: https://github.com/rs/zerolog?tab=readme-ov-file#multiple-log-output
//
// loggerOpts: is of type *LoggerOpts, if nil, default options are used
// to enable step level logging, set StepLoggingEnabled to true
func NewGoStepsLogger(out io.Writer, loggerOpts *LoggerOpts) goStepsLogger {
	zerolog.TimeFieldFormat = zerolog.TimeFormatUnix

	if out == nil {
		out = defaultLogOut
	}

	if loggerOpts == nil {
		loggerOpts = &LoggerOpts{
			StepLoggingEnabled: false,
		}
	}

	return goStepsLogger{
		config: loggerOpts,
		logger: zerolog.New(out).With().Timestamp().Logger(),
	}
}

type stepLogStruct struct {
	Name     string
	State    string
	Message  *string
	Error    error
	RunCount int
	MaxRun   int
}

// getStepLogStruct returns the loggable struct for the step
func (step *Step) getStepLogStruct() stepLogStruct {
	var stepError error
	if step.stepResult.StepError != nil {
		stepError = step.stepResult.StepError
	}

	return stepLogStruct{
		Name: string(step.Name),

		State:   string(step.stepResult.StepState),
		Message: step.stepResult.StepMessage,
		Error:   stepError,

		RunCount: step.stepRunProgress.runCount,
		MaxRun:   step.StepOpts.MaxRunAttempts,
	}
}

// loggableFormat returns the loggable format for the step
func (s stepLogStruct) loggableFormat() map[string]interface{} {
	loggableFields := map[string]interface{}{
		"maxRun": s.MaxRun,
	}

	if s.Error != nil {
		loggableFields["error"] = s.Error
	}

	if s.Message != nil {
		loggableFields["message"] = s.Message
	}

	return loggableFields
}

// log logs the step with the step name, state, run count and the log fields
// it is only used by the step if the step logging is enabled
func (c *GoStepsCtx) log(step *Step) {
	lStruct := step.getStepLogStruct()

	c.logger.logger.WithLevel(
		stateToLevelMap[step.stepResult.StepState],
	).Str(
		"step", string(lStruct.Name),
	).Str(
		"state", lStruct.State,
	).Int(
		"runCount", lStruct.RunCount,
	).Fields(
		lStruct.loggableFormat(),
	).Msg("")
}

// Log logs the message with the step name and the log level, if provided.
// The log level is the first argument in the args.
func (c *GoStepsCtx) Log(message string, levels ...LogLevel) {
	ll := InfoLevel
	if len(levels) > 0 {
		ll = levels[0]
	}

	c.logger.logger.WithLevel(zerolog.Level(ll)).Str(
		"step", string(c.currentStep),
	).Msg(message)
}
