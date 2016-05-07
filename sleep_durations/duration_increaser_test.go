package sleep_durations

import (
	"fmt"
	"testing"
	"time"

	_ "github.com/go-sql-driver/mysql"
	. "github.com/smartystreets/goconvey/convey"
)

func TestDurationIncreaser(t *testing.T) {
	Convey("Testing DurationIncreaser", t, func() {
		iterationsPerDuration := 4
		durationList := []time.Duration{
			500 * time.Millisecond,
			2 * time.Second,
			10 * time.Second,
			30 * time.Second,
			1 * time.Minute,
		}

		d := New(iterationsPerDuration, durationList)

		expectedDurations := []time.Duration{
			500 * time.Millisecond, 500 * time.Millisecond, 500 * time.Millisecond, 500 * time.Millisecond,
			2 * time.Second, 2 * time.Second, 2 * time.Second, 2 * time.Second,
			10 * time.Second, 10 * time.Second, 10 * time.Second, 10 * time.Second,
			30 * time.Second, 30 * time.Second, 30 * time.Second, 30 * time.Second,
			1 * time.Minute, 1 * time.Minute, 1 * time.Minute, 1 * time.Minute,
		}
		lastExpectedStickyDuration := expectedDurations[len(expectedDurations)-1]

		actualDurations := []time.Duration{}
		for _ = range expectedDurations {
			actualDurations = append(actualDurations, d.Next())
		}
		So(actualDurations, ShouldResemble, expectedDurations)

		for i := 0; i < 1000; i++ {
			actual := d.Next()
			if actual != lastExpectedStickyDuration {
				So(fmt.Errorf("Actual duration '%s' does not equal expected '%s' (i = %d)", actual.String(), lastExpectedStickyDuration.String(), i), ShouldBeNil)
			} else {
				So(d.Next(), ShouldEqual, lastExpectedStickyDuration)
			}
		}
	})
}
