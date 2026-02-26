package storage

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// DB содержит пул соединений к PostgreSQL
type DB struct {
	Pool *pgxpool.Pool
}

// Connect инициализирует подключение к базе данных.
// Использует пул соединений (pgxpool) для эффективной обработки конкурентных запросов
// от REST API и фонового движка автоматизации.
func Connect(ctx context.Context) (*DB, error) {
	// Собираем строку подключения из переменных окружения
	host := os.Getenv("DB_HOST")
	port := os.Getenv("DB_PORT")
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")

	// Если переменные не заданы (например, забыли .env), ставим дефолт
	if host == "" {
		host = "localhost"
	}
	if port == "" {
		port = "5432"
	}
	if user == "" {
		user = "terrarium"
	}
	if dbname == "" {
		dbname = "terrarium_db"
	}

	dsn := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable", user, password, host, port, dbname)

	config, err := pgxpool.ParseConfig(dsn)
	if err != nil {
		return nil, fmt.Errorf("ошибка парсинга конфигурации БД: %w", err)
	}

	// Настройки пула
	config.MaxConns = 10
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.MaxConnIdleTime = 30 * time.Minute

	log.Printf("Подключение к PostgreSQL на %s:%s...", host, port)

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("не удалось создать пул соединений: %w", err)
	}

	// Проверяем, что база реально отвечает
	if err := pool.Ping(ctx); err != nil {
		return nil, fmt.Errorf("база данных не отвечает на ping: %w", err)
	}

	log.Println("Успешное подключение к PostgreSQL!")

	return &DB{Pool: pool}, nil
}

// Close закрывает все открытые соединения с базой
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		log.Println("Соединение с PostgreSQL закрыто.")
	}
}
