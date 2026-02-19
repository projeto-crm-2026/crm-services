package repo

import (
	"context"
	_ "embed"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed seeds/seed.sql
var seedSQL string

type Repo interface {
	User() UserRepo
}

type Conn struct {
	db      *gorm.DB
	pgxPool *pgxpool.Pool
}

func Connect(ctx context.Context, dbConfig config.DBConfig) (*Conn, error) {
	// only for migrations
	gormDSN := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=disable Timezone=America/Sao_Paulo search_path=public",
		dbConfig.Address, dbConfig.User, dbConfig.Password, dbConfig.Name, dbConfig.Port)

	db, err := gorm.Open(postgres.Open(gormDSN), &gorm.Config{})
	if err != nil {
		return nil, fmt.Errorf("failed to connect to database with GORM: %w", err)
	}

	pgxDSN := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable&search_path=public",
		dbConfig.User, dbConfig.Password, dbConfig.Address, dbConfig.Port, dbConfig.Name)

	pgxPool, err := pgxpool.New(ctx, pgxDSN)
	if err != nil {
		return nil, fmt.Errorf("failed to create pgx pool: %w", err)
	}

	if err := pgxPool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	return &Conn{
		db:      db,
		pgxPool: pgxPool,
	}, nil
}

func RunCustomMigrations(db *gorm.DB) error {
	if err := db.Exec(`
        CREATE INDEX IF NOT EXISTS idx_webhook_events_gin 
        ON webhook USING GIN (events);
    `).Error; err != nil {
		return fmt.Errorf("failed to run custom migration: %w", err)
	}

	return nil
}

func RunSeeds(db *gorm.DB) error {
	if err := db.Exec(seedSQL).Error; err != nil {
		return fmt.Errorf("failed to run seed: %w", err)
	}
	return nil
}

func (c *Conn) GetDB() *gorm.DB {
	return c.db
}

func (c *Conn) GetPool() *pgxpool.Pool {
	return c.pgxPool
}

func (c *Conn) Close() {
	if c.pgxPool != nil {
		c.pgxPool.Close()
	}
}
