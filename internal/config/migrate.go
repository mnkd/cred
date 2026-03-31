package config

import "fmt"

// currentVersion は現在の設定ファイルスキーマバージョン
const currentVersion = 1

// migration はバージョン N から N+1 への変換関数
type migration func(*Config) error

// migrations はバージョン番号をキーとするマイグレーション関数のマップ。
// バージョン N のマイグレーションは N → N+1 への変換を行う。
var migrations = map[int]migration{
	// NOTE: 将来のスキーマ変更時にここへ追加する
	// 例:
	// 1: migrateV1toV2,
}

// Migrate は設定ファイルのバージョンが古い場合に順次マイグレーションを実行する。
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
