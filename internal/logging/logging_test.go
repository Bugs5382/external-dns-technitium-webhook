package logging

/*
Apache License 2.0

Copyright 2026 external-dns-technitium-webhook Contributors

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

import (
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

// LOG_LEVEL accepts zerolog names and their numeric values. Anything outside
// zerolog's range must fall back to info instead of wrapping around int8
// (gosec G115, issue #28).
func TestSetLogLevel(t *testing.T) {
	cases := []struct {
		in   string
		want zerolog.Level
	}{
		{"", zerolog.InfoLevel},
		{"debug", zerolog.DebugLevel},
		{"TRACE", zerolog.TraceLevel},
		{"-1", zerolog.TraceLevel},
		{"3", zerolog.ErrorLevel},
		{"7", zerolog.Disabled},
		{"44", zerolog.InfoLevel},
		{"300", zerolog.InfoLevel},
		{"-200", zerolog.InfoLevel},
		{"bogus", zerolog.InfoLevel},
	}
	prev := zerolog.GlobalLevel()
	t.Cleanup(func() { zerolog.SetGlobalLevel(prev) })

	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			t.Setenv("LOG_LEVEL", tc.in)
			zerolog.SetGlobalLevel(zerolog.PanicLevel)
			setLogLevel()
			assert.Equal(t, tc.want, zerolog.GlobalLevel())
		})
	}
}
