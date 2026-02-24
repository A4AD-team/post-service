package util

import "math"

func CalculateHotScore(likes, comments, views int64) float64 {
	return float64(likes)*0.8 + float64(comments)*0.5 + math.Log(float64(views)+1)
}
