package technitium

import "time"

func sessionBuffer() time.Duration {
	config := StartupConfig{}
	return time.Duration(config.SessionTTL) * time.Minute
}
