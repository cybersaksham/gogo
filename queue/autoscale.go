package queue

import "time"

type AutoscaleConfig struct {
	MinConcurrency    int
	MaxConcurrency    int
	ScaleUpReadyTasks int
	ScaleDownIdleFor  time.Duration
}

type AutoscaleState struct {
	MinConcurrency    int
	MaxConcurrency    int
	ScaleUpReadyTasks int
	ScaleDownIdleFor  time.Duration
}

func ResolveAutoscale(baseConcurrency int, config AutoscaleConfig) AutoscaleState {
	if baseConcurrency < 1 {
		baseConcurrency = 1
	}
	minimum := config.MinConcurrency
	if minimum < 1 {
		minimum = baseConcurrency
	}
	maximum := config.MaxConcurrency
	if maximum < minimum {
		maximum = minimum
	}
	scaleUpAt := config.ScaleUpReadyTasks
	if scaleUpAt < 1 {
		scaleUpAt = maximum
	}
	return AutoscaleState{
		MinConcurrency:    minimum,
		MaxConcurrency:    maximum,
		ScaleUpReadyTasks: scaleUpAt,
		ScaleDownIdleFor:  config.ScaleDownIdleFor,
	}
}

func (s AutoscaleState) Target(readyTasks int) int {
	if readyTasks >= s.ScaleUpReadyTasks {
		return s.MaxConcurrency
	}
	return s.MinConcurrency
}
