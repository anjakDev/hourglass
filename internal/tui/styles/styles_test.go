package styles_test

import (
	"testing"
	"time"

	"github.com/anjakDev/hourglass/internal/tui/styles"
	"github.com/stretchr/testify/assert"
)

func TestFormatDuration(t *testing.T) {
	cases := []struct {
		d    time.Duration
		want string
	}{
		{0, "0m"},
		{30 * time.Second, "0m"},
		{time.Minute, "1m"},
		{45 * time.Minute, "45m"},
		{time.Hour, "1h 0m"},
		{time.Hour + 23*time.Minute, "1h 23m"},
		{2*time.Hour + time.Minute, "2h 1m"},
		{10*time.Hour + 59*time.Minute, "10h 59m"},
	}
	for _, tc := range cases {
		assert.Equal(t, tc.want, styles.FormatDuration(tc.d), "FormatDuration(%v)", tc.d)
	}
}
