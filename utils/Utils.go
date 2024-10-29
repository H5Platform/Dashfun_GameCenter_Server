package utils

import "time"

func IsSameDay(t1, t2 time.Time) bool {
	t1d := t1.YearDay()
	t1y := t1.Year()
	t2d := t2.YearDay()
	t2y := t2.Year()
	return t1y != t2y || t1d != t2d
}
