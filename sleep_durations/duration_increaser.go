package sleep_durations

import "time"

//DurationIncreaser is used to increase time.Duration the longer its running
type DurationIncreaser interface {
	Next() time.Duration
}

//New creates a new instance of DurationIncreaser
func New(iterationsPerDuration int, durationList []time.Duration) DurationIncreaser {
	return &durationIncreaser{
		IterationsPerDuration: iterationsPerDuration,
		DurationList:          durationList,
	}
}

type durationIncreaser struct {
	iteration int
	listIndex int

	IterationsPerDuration int
	DurationList          []time.Duration
}

func (d *durationIncreaser) inc() {
	d.iteration++
	if d.iteration >= d.IterationsPerDuration {
		d.iteration = 0
		if d.listIndex < len(d.DurationList)-1 {
			d.listIndex++
		}
	}
}

func (d *durationIncreaser) Next() time.Duration {
	defer d.inc()
	return d.DurationList[d.listIndex]
}
