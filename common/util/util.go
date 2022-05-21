package util

import (
	"math"
)

func Reverse[K any](arr []K) {
	length := len(arr)
	for i, j := 0, length-1; i < j; i, j = i+1, j-1 {
		v := arr[i]
		arr[i] = arr[j]
		arr[j] = v
	}
}

func MaxInt(arr ...int) int {
	maxValue := math.MinInt32

	for _, val := range arr {
		if val > maxValue {
			maxValue = val
		}
	}

	return maxValue
}
