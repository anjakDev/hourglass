package reports

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestSparkline_Empty(t *testing.T) {
	assert.Equal(t, "", sparkline(nil))
}

func TestSparkline_AllZero(t *testing.T) {
	values := []time.Duration{0, 0, 0, 0, 0, 0, 0}
	assert.Equal(t, "▁▁▁▁▁▁▁", sparkline(values))
}

func TestSparkline_SingleMax(t *testing.T) {
	values := []time.Duration{0, 0, 0, time.Hour, 0, 0, 0}
	assert.Equal(t, "▁▁▁█▁▁▁", sparkline(values))
}

func TestSparkline_FullGradient(t *testing.T) {
	values := []time.Duration{
		0, time.Hour, 2 * time.Hour, 3 * time.Hour,
		4 * time.Hour, 5 * time.Hour, 6 * time.Hour, 7 * time.Hour,
	}
	assert.Equal(t, "▁▂▃▄▅▆▇█", sparkline(values))
}

func TestSparkline_AllSame(t *testing.T) {
	values := []time.Duration{time.Hour, time.Hour, time.Hour}
	assert.Equal(t, "███", sparkline(values))
}
