package config

import "fmt"

// currentVersion is the current config file schema version.
const currentVersion = 1

// migration is a function that converts from version N to N+1.
type migration func(*Config) error

// migrations maps version numbers to their migration functions.
// A migration for version N converts from N to N+1.
var migrations = map[int]migration{
	// NOTE: Add future schema migrations here
	// e.g.:
	// 1: migrateV1toV2,
}

// Migrate runs sequential migrations if the config file version is outdated.
func Migrate(cfg *Config) (bool, error) {
	if cfg.Version >= currentVersion {
		return false, nil
	}

	if cfg.Version == 0 {
		cfg.Version = 1
	}

	migrated := false
	for cfg.Version < currentVersion {
		fn, ok := migrations[cfg.Version]
		if !ok {
			return migrated, fmt.Errorf("no migration path from version %d", cfg.Version)
		}
		if err := fn(cfg); err != nil {
			return migrated, fmt.Errorf("migration from version %d failed: %w", cfg.Version, err)
		}
		cfg.Version++
		migrated = true
	}

	return migrated, nil
}
