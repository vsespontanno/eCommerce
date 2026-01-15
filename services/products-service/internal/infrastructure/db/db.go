package db

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"go.uber.org/zap"
)

func ConnectToPostgres(user, password, dbname, host, port string, logger *zap.SugaredLogger) (*sqlx.DB, error) {
	// Формируем строку подключения
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		user, password, dbname, host, port)

	// Открываем соединение с базой данных
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Проверяем соединение
	err = db.Ping()
	if err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(25)                 // Максимальное число открытых соединений
	db.SetMaxIdleConns(5)                  // Максимальное число бездействующих соединений
	db.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни соединения
	db.SetConnMaxIdleTime(2 * time.Minute) // Максимальное время бездействия соединения

	logger.Infow("Connected to Postgres",
		"host", host,
		"port", port,
		"database", dbname,
		"max_open_conns", 25,
		"max_idle_conns", 5,
	)
	return db, nil
}
