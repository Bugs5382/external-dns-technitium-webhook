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
	"os"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

func Init() {
	setLogLevel()
	setLogFormat()
}

func setLogFormat() {
	if strings.ToLower(os.Getenv("LOG_FORMAT")) == "json" {
		return
	}
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})
}

func setLogLevel() {
	levelStr := os.Getenv("LOG_LEVEL")
	if levelStr == "" {
		zerolog.SetGlobalLevel(zerolog.InfoLevel)
		return
	}
	// zerolog.ParseLevel accepts names and any int8, so an in-range number such
	// as 44 still parses. Only accept real levels, trace (-1) to disabled (7);
	// anything else falls back to info instead of wrapping (gosec G115, #28).
	parsedLevel, err := zerolog.ParseLevel(levelStr)
	if err == nil && parsedLevel >= zerolog.TraceLevel && parsedLevel <= zerolog.Disabled {
		zerolog.SetGlobalLevel(parsedLevel)
		return
	}
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Warn().Str("LOG_LEVEL", levelStr).Msg("invalid log level, defaulting to info")
}
