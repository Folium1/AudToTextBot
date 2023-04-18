package config

import (
	"context"

	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// RedisConfig represents the configuration for the Redis server
type RedisConfig struct {
	Addr     string
	Password string
	DB       int
}

// NewRedisConfig creates a new Redis configuration object from environment variables
func NewRedisConfig() *RedisConfig {
	err := godotenv.Load()
	if err != nil {
		log.Fatalf("Error loading .env file: %v", err)
	}

	return &RedisConfig{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	}
}

// ConnectToRedis connects to the Redis server using the configuration parameters
func (c *RedisConfig) ConnectToRedis() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.Addr,
		Password: c.Password,
		DB:       c.DB,
	})

	// Check if the Redis server is responsive
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %v", err)
	}

	log.Printf("Connected to Redis at %s:%d", c.Addr, c.DB)
	return redisClient, nil
}

// RedisServiceConfig represents the configuration for the Redis service
type RedisServiceConfig struct {
	RedisConfig *RedisConfig
}

// NewRedisServiceConfig creates a new Redis service configuration object from environment variables
func NewRedisServiceConfig() *RedisServiceConfig {
	return &RedisServiceConfig{
		RedisConfig: NewRedisConfig(),
	}
}

// NewRedisClient initializes a new Redis client using the Redis configuration parameters
func (c *RedisServiceConfig) NewRedisClient() (*redis.Client, error) {
	redisClient := redis.NewClient(&redis.Options{
		Addr:     c.RedisConfig.Addr,
		Password: c.RedisConfig.Password,
		DB:       c.RedisConfig.DB,
	})

	// Check if the Redis server is responsive
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := redisClient.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to ping Redis: %v", err)
	}

	log.Printf("Connected to Redis at %s:%d", c.RedisConfig.Addr, c.RedisConfig.DB)
	return redisClient, nil
}
