package seed

import (
	"fmt"
	"os"
	"strings"
)

// seedableEnvs lists the APP_ENV values where seeding is allowed. The check is an
// allowlist rather than a "production" denylist: a missing or misspelled APP_ENV
// must never be enough to truncate a real database.
var seedableEnvs = map[string]bool{
	"dev":         true,
	"development": true,
	"local":       true,
	"test":        true,
	"staging":     true,
}

// refuseSeedUnlessSeedableEnv guards both Run (truncate + demo data) and RunMass.
func refuseSeedUnlessSeedableEnv(cmd string) error {
	env := strings.ToLower(strings.TrimSpace(os.Getenv("APP_ENV")))
	if seedableEnvs[env] {
		return nil
	}
	if env == "" {
		return fmt.Errorf("%s refused: APP_ENV is not set (expected one of dev, local, test, staging)", cmd)
	}
	return fmt.Errorf("%s refused for APP_ENV=%q (expected one of dev, local, test, staging)", cmd, env)
}
