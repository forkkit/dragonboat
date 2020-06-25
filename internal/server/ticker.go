// Copyright 2017-2019 Lei Ni (nilei81@gmail.com) and other Dragonboat authors.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package server

import (
	"time"
)

// TickerFunc is type of the function that will be called by the RunTicker
// function after each tick. The returned boolean value indicates whether the
// ticker should stop.
type TickerFunc func(usec uint64) bool

// StartTicker runs a ticker at the specified interval, the provided TickerFunc
// will be called after each tick. The ticker will be stopped when the
// TickerFunc return a true value or when any of the two specified stop
// channels is signalled.
func StartTicker(td time.Duration, tf TickerFunc, stopc <-chan struct{}) {
	// FIXME: use Milliseconds() once go1.14 is released
	tms := td.Nanoseconds() / 1000000
	if tms == 0 {
		panic("invalid duration")
	}
	if tms == 1 {
		run1MSTicker(tf, stopc)
	} else {
		runLFTicker(td, tf, stopc)
	}
}

func run1MSTicker(tf TickerFunc, stopc <-chan struct{}) {
	ticker := time.NewTicker(time.Millisecond)
	defer ticker.Stop()
	count := 0
	for range ticker.C {
		count++
		if count%10 == 0 {
			select {
			case <-stopc:
				return
			default:
			}
		}
		if tf(1000) {
			return
		}
	}
}

func runLFTicker(td time.Duration, tf TickerFunc, stopc <-chan struct{}) {
	ticker := time.NewTicker(td)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			// FIXME: use Microseconds() once go1.14 is released
			if tf(uint64(td.Nanoseconds() / 1000)) {
				return
			}
		case <-stopc:
			return
		}
	}
}
