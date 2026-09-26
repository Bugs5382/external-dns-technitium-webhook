package server

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
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

// The health and metrics listener must bound how long a client may take to
// send its headers, so slow clients cannot hold connections open (gosec G112,
// issue #28).
func TestNewHealthServer_SetsTimeouts(t *testing.T) {
	s := newHealthServer("127.0.0.1:0", http.NewServeMux())

	assert.Equal(t, "127.0.0.1:0", s.Addr)
	assert.Positive(t, s.ReadHeaderTimeout)
	assert.Positive(t, s.ReadTimeout)
	assert.Positive(t, s.WriteTimeout)
	assert.Positive(t, s.IdleTimeout)
}
