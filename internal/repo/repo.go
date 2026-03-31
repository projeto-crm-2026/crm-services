package repo

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/projeto-crm-2026/crm-services/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

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
