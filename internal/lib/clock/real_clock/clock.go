package real_clock

import "time"

type RealClock struct {
}

func NewRealClock() *RealClock {
	return &RealClock{}
}

func (r *RealClock) Now() time.Time {
	return time.Now()
}
