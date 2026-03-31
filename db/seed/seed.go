package seed

import (
	_ "embed"
	"fmt"

	"gorm.io/gorm"
)

//go:embed seed.sql
var seedSQL string

func RunSeeds(db *gorm.DB) error {
	if err := db.Exec(seedSQL).Error; err != nil {
		return fmt.Errorf("failed to run seed: %w", err)
	}
	return nil
}
