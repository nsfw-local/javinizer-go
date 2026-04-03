package migrations

import "github.com/javinizer/javinizer-go/internal/config"

func init() {
	config.RegisterMigration(config.NewLegacyMigration())
}
