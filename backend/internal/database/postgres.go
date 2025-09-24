package database

import (
	"context"
	"database/sql"
	"log"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

func InitDB(databaseURL string) *sql.DB {
	db, err := sql.Open("postgres", databaseURL)
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	// Configure connection pool
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Test the connection
	if err := db.Ping(); err != nil {
		log.Fatal("Failed to ping database:", err)
	}

	log.Println("Successfully connected to PostgreSQL database")
	return db
}

func InitRedis(redisURL string) *redis.Client {
	rdb := redis.NewClient(&redis.Options{
		Addr:     strings.TrimPrefix(redisURL, "redis://"),
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	// Test the connection
	if err := rdb.Ping(context.Background()).Err(); err != nil {
		log.Fatal("Failed to connect to Redis:", err)
	}

	log.Println("Successfully connected to Redis")
	return rdb
}