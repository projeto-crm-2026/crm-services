package repo

import (
	"fmt"

	"github.com/projeto-crm-2026/crm-services/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

type Repo interface {
	User() UserRepo
}

type Conn struct {
	db *gorm.DB
}

func Connect(dbConfig config.DBConfig) (*Conn, error) {
	dns := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable Timezone=America/Sao_Paulo search_path=public",
		dbConfig.Address, dbConfig.User, dbConfig.Password, dbConfig.Name, dbConfig.Port)
	db, err := gorm.Open(postgres.Open(dns), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database: %w", err)
	}

	return &Conn{db: db}, nil
}

func (c *Conn) GetDB() *gorm.DB {
	return c.db
}
