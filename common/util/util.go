package util

import (
	"math"
	"reflect"
)

func Reverse(arr interface{}) {
	if arr != nil {
		length := reflect.ValueOf(arr).Len()
		swap := reflect.Swapper(arr)
		for i, j := 0, length-1; i < j; i, j = i+1, j-1 {
			swap(i, j)
		}
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
