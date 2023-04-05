package config

import (
	"context"
	"database/sql"

	"fmt"
	"log"
	"os"
	"time"

	_ "github.com/go-sql-driver/mysql"
	"github.com/joho/godotenv"
	"github.com/redis/go-redis/v9"
)

// MySQLConfig represents the configuration for the MySQL database
type MySQLConfig struct {
	DBHost     string
	DBPort     string
	DBUser     string
	DBPassword string
	DBName     string
}

// NewMySQLConfig creates a new MySQL configuration object from environment variables
func NewMysqlConfig() *MySQLConfig {
	return &MySQLConfig{
		DBHost:     os.Getenv("MYSQL_HOST"),
		DBPort:     os.Getenv("MYSQL_PORT"),
		DBUser:     os.Getenv("MYSQL_USER"),
		DBPassword: os.Getenv("MYSQL_PASSWORD"),
		DBName:     os.Getenv("MYSQL_NAME"),
	}
}

// ConnectToMySQL connects to the MySQL database using the instance parameters
func (c *MySQLConfig) ConnectToMySQL() (*sql.DB, error) {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s", c.DBUser, c.DBPassword, c.DBHost, c.DBPort, c.DBName)

	db, err := sql.Open("mysql", dsn)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to MySQL: %v", err)
	}
	defer db.Close()

	// Check if the database is responsive
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	if err := db.PingContext(ctx); err != nil {
		return nil, fmt.Errorf("failed to ping MySQL: %v", err)
	}

	_, err = db.Exec("CREATE TABLE scripts(file_name VARCHAR(100),duration INT,user_id BIGINT,script TEXT(10000));")
	if err != nil {
		return nil, fmt.Errorf("failed to create scripts table: %v", err)
	}

	log.Printf("Connected to MySQL at %s:%s/%s", c.DBHost, c.DBPort, c.DBName)
	return db, nil
}

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
