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
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
)

// Init sets up logging configuration
func Init() {
	setLogLevel()
	setLogFormat()
}

func setLogFormat() {
	format := strings.ToLower(os.Getenv("LOG_FORMAT"))

	if format == "json" {
		return
	}

	// Otherwise, default to ConsoleWriter (Text)
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

	parsedLevel, err := zerolog.ParseLevel(levelStr)
	if err == nil {
		zerolog.SetGlobalLevel(parsedLevel)
		return
	}

	// 3. Fallback: Try parsing as an integer (e.g., "0" for debug, "1" for info)
	if levelInt, err := strconv.Atoi(levelStr); err == nil {
		zerolog.SetGlobalLevel(zerolog.Level(levelInt))
		return
	}

	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	log.Debug().Msgf("Invalid log level '%s', defaulting to info", levelStr)
}
