package statCalculator

import (
	"math"
)

func CalculateHp(base, iv, ev, level int) int {
	inner := ((2 * base + iv + int(math.Floor(float64(ev / 4)))) * level) / 100
	hpStat := int(math.Floor(float64(inner))) + level + 10
	return hpStat
}
func CalculateOtherStat(base, iv, ev, level int, nature float64) int {
	inner := (2 * base + iv + int(math.Floor(float64(ev / 4)))) * level
	otherStat := int(math.Floor((math.Floor(float64(inner / 100)) + 5) * nature))
	return otherStat
}

