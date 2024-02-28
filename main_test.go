package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"
)

func Test_parseTime(t *testing.T) {
	testCases := []struct {
		s      string
		err    bool
		hour   int
		minute int
	}{
		{"2:45PM", false, 14, 45},
		{"2:45pm", false, 14, 45},
		{"13:37", false, 13, 37},
		{"4pm", false, 16, 00},
		{"6", true, 0, 0},
	}

	for _, tc := range testCases {
		t.Run(tc.s, func(t *testing.T) {
			tm, err := parseTime(tc.s)
			if tc.err {
				require.Error(t, err)
				return
			}

			now := time.Now()
			require.Equal(t, tc.hour, tm.Hour())
			require.Equal(t, tc.minute, tm.Minute())

			require.Equal(t, now.Year(), tm.Year())
			require.Equal(t, now.Month(), tm.Month())
			require.Equal(t, now.Day(), tm.Day())
		})
	}
}
