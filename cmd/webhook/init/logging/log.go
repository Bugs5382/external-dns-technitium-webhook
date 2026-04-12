package logging

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
