package db

import (
	"fmt"
	"time"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/vsespontanno/eCommerce/pkg/logger"
)

func ConnectToPostgres(user, password, dbname, host, port string) (*sqlx.DB, error) {
	// Формируем строку подключения
	connStr := fmt.Sprintf("user=%s password=%s dbname=%s host=%s port=%s sslmode=disable",
		user, password, dbname, host, port)

	// Открываем соединение с базой данных
	db, err := sqlx.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database connection: %w", err)
	}

	// Проверяем соединение с retry
	maxRetries := 10
	for i := 0; i < maxRetries; i++ {
		err = db.Ping()
		if err == nil {
			break
		}
		logger.Log.Warnw("Failed to ping database, retrying...",
			"attempt", i+1,
			"max_retries", maxRetries,
			"error", err,
		)
		time.Sleep(100 * time.Millisecond)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to ping database after %d retries: %w", maxRetries, err)
	}

	// Настраиваем пул соединений
	db.SetMaxOpenConns(25)                 // Максимальное число открытых соединений
	db.SetMaxIdleConns(5)                  // Максимальное число бездействующих соединений
	db.SetConnMaxLifetime(5 * time.Minute) // Максимальное время жизни соединения
	db.SetConnMaxIdleTime(2 * time.Minute) // Максимальное время бездействия соединения

	logger.Log.Info("Connected to Postgres")
	return db, nil
}
