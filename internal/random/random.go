package random

import (
	"math"
	"math/rand"
	"time"
)

func RandNameItem(min int, max int) string {
	randStatus := RandInt(min, max)
	if randStatus == 2 {
		return "My table"
	}
	return "You table"
}

func RandInt(min int, max int) int {
	rand.Seed(time.Now().UnixNano())
	return 2 + rand.Intn(max-min+(min-1))
}

func RandPriceItem(min float64, max float64) float64 {
	randVal := min + rand.Float64()*(max-min)
	return math.Ceil(randVal*100) / 100
}
