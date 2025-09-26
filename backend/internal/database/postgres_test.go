package database

import (
	"context"
	"os"
	"testing"
	"time"
)

var ctx = context.Background()

func TestInitDB(t *testing.T) {
	// Skip test if DATABASE_URL is not set
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL environment variable not set, skipping database test")
	}

	// Test database connection
	db := InitDB(databaseURL)
	if db == nil {
		t.Fatal("Failed to initialize database connection")
	}
	defer db.Close()

	// Test database ping
	if err := db.Ping(); err != nil {
		t.Fatalf("Failed to ping database: %v", err)
	}

	// Test connection pool settings
	stats := db.Stats()
	if stats.OpenConnections > 25 {
		t.Errorf("Too many open connections: %d", stats.OpenConnections)
	}
}

func TestInitRedis(t *testing.T) {
	// Skip test if REDIS_URL is not set
	redisURL := os.Getenv("REDIS_URL")
	if redisURL == "" {
		t.Skip("REDIS_URL environment variable not set, skipping Redis test")
	}

	// Test Redis connection
	rdb := InitRedis(redisURL)
	if rdb == nil {
		t.Fatal("Failed to initialize Redis connection")
	}
	defer rdb.Close()

	// Test Redis ping
	if err := rdb.Ping(ctx).Err(); err != nil {
		t.Fatalf("Failed to ping Redis: %v", err)
	}
}

func TestDatabaseConnectionPool(t *testing.T) {
	// Skip test if DATABASE_URL is not set
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		t.Skip("DATABASE_URL environment variable not set, skipping database test")
	}

	db := InitDB(databaseURL)
	defer db.Close()

	// Test that connection pool is configured correctly
	if err := db.Ping(); err != nil {
		t.Fatalf("Database ping failed: %v", err)
	}

	// Test multiple connections
	for i := 0; i < 5; i++ {
		go func() {
			if err := db.Ping(); err != nil {
				t.Errorf("Concurrent database ping failed: %v", err)
			}
		}()
	}

	// Wait a bit for goroutines to complete
	time.Sleep(100 * time.Millisecond)

	// Check final stats
	stats := db.Stats()
	if stats.OpenConnections > 25 {
		t.Errorf("Too many open connections after concurrent pings: %d", stats.OpenConnections)
	}
}

