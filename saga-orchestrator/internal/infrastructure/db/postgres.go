package db

import (
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/vsespontanno/eCommerce/saga-orchestrator/internal/config"
	"go.uber.org/zap"
)

func NewPostgresDB(cfg *config.Config, log *zap.SugaredLogger) (*sqlx.DB, error) {
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=disable",
		cfg.PGHost,
		cfg.PGPort,
		cfg.PGUser,
		cfg.PGPassword,
		cfg.PGName,
	)

	db, err := sqlx.Connect("postgres", dsn)
	if err != nil {
		log.Errorw("failed to connect to postgres", "error", err)
		return nil, err
	}

	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(5)

	if err := db.Ping(); err != nil {
		log.Errorw("failed to ping postgres", "error", err)
		return nil, err
	}

	log.Infow("successfully connected to postgres",
		"host", cfg.PGHost,
		"port", cfg.PGPort,
		"database", cfg.PGName,
	)

	return db, nil
}
