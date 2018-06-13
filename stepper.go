package main

import "math"

func GetStep(val float64, steps []float64) int {
	bestI := 0
	bestDiff := math.Inf(1)
	for i, step := range steps {
		diff := math.Abs(val - step)
		if diff < bestDiff {
			bestDiff = diff
			bestI = i
		}
	}
	return bestI
}

func NextStep(val float64, steps []float64) float64 {
	i := GetStep(val, steps)
	i += 1
	i %= len(steps)
	return steps[i]
}

func PrevStep(val float64, steps []float64) float64 {
	i := GetStep(val, steps)
	i += len(steps) - 1
	i %= len(steps)
	return steps[i]
}
