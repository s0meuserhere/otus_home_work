package event

import (
	"time"
)

// TimeRangesIntersect проверяет пересечение интервалов [aStart, aEnd) и [bStart, bEnd).
func TimeRangesIntersect(aStart, aEnd, bStart, bEnd time.Time) bool {
	return aStart.Before(bEnd) && bStart.Before(aEnd)
}
