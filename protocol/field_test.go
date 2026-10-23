// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestDurationStrParse(t *testing.T) {
	t.Run("valid", func(t *testing.T) {
		cases := []struct {
			name string
			in   DurationStr
			want time.Duration
		}{
			{"seconds", "30s", 30 * time.Second},
			{"minutes", "5m", 5 * time.Minute},
			{"hours", "2h", 2 * time.Hour},
			{"millis", "150ms", 150 * time.Millisecond},
			{"composite", "1h30m", 90 * time.Minute},
			{"negative", "-10s", -10 * time.Second},
			{"zero", "0s", 0},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				got, err := c.in.Parse()
				require.NoError(t, err)
				assert.Equal(t, c.want, got)
			})
		}
	})

	t.Run("invalid", func(t *testing.T) {
		cases := []struct {
			name string
			in   DurationStr
		}{
			{"empty", ""},
			{"no unit", "10"},
			{"unknown unit", "10x"},
			{"garbage", "abc"},
		}
		for _, c := range cases {
			t.Run(c.name, func(t *testing.T) {
				_, err := c.in.Parse()
				assert.Error(t, err)
			})
		}
	})
}
