package gosteps

import (
	"math"
)

const (
	// to avoid infinite runs due to the MaxAttempts not being set, we're keeping the default attempts to 100
	// if required, import and use the MaxMaxAttempts in the step.MaxAttempts field
	DefaultMaxAttempts = 100

	// the Max value is 9223372036854775807, which is not infinite but a huge number of attempts
	MaxMaxAttempts = math.MaxInt
)