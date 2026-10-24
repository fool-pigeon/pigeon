// Copyright (c) 2026 Homoarea <homoarea114@gmail.com>
//
// This software is released under the MIT License.
// https://opensource.org/licenses/MIT

package protocol

import "time"

type DurationStr string

func (ts DurationStr) Parse() (time.Duration, error) {
	if ts == "" {
		return 0, nil
	}
	return time.ParseDuration(string(ts))
}
