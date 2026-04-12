package configuration

import (
	"time"

	"github.com/caarlos0/env/v11"
	log "github.com/sirupsen/logrus"
)

// Config struct for configuration environmental variables
type Config struct {
	ServerHost           string        `env:"SERVER_HOST" envDefault:"127.0.0.1"`
	ServerPort           int           `env:"SERVER_PORT" envDefault:"8888"`
	HealthCheckPort      int           `env:"HEALTH_CHECK_PORT" envDefault:"8080"`
	ServerReadTimeout    time.Duration `env:"SERVER_READ_TIMEOUT"`
	ServerWriteTimeout   time.Duration `env:"SERVER_WRITE_TIMEOUT"`
	DomainFilter         []string      `env:"DOMAIN_FILTER" envDefault:""`
	ExcludeDomains       []string      `env:"EXCLUDE_DOMAIN_FILTER" envDefault:""`
	RegexDomainFilter    string        `env:"REGEXP_DOMAIN_FILTER" envDefault:""`
	RegexDomainExclusion string        `env:"REGEXP_DOMAIN_FILTER_EXCLUSION" envDefault:""`
	RegexNameFilter      string        `env:"REGEXP_NAME_FILTER" envDefault:""`
}

// Init sets up configuration by reading set environmental variables
func Init() Config {
	cfg := Config{}
	if err := env.Parse(&cfg); err != nil {
		log.Fatalf("Error reading configuration from environment: %v", err)
	}
	return cfg
}
