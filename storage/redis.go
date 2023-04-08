package storage

import (
	"context"
	"fmt"
	"log"
	"tgbot/config"
	"time"

	"github.com/redis/go-redis/v9"
)

var (
	monthDuration = 24 * time.Hour * 30
	ctx           = context.Background()
)

var RedisService interface {
	SavePremiumUser(userId int) error
	IncrementPremiumTime(userId, duration int) (int, error)
	GetPremiumTime(userId int) (int, error)
	IsPremium(userId int) (int, error)
	GetUnpremiumTries(userId int) (int, error)
	SaveUnpremiumUser(userId int) error
	IcrementUnpmremiumTries(userId int) (int, error)
}

type RedisStorage struct {
	db *redis.Client
}

func NewStorage() (*RedisStorage, error) {
	redisConn := config.NewRedisConfig()
	client, err := redisConn.ConnectToRedis()
	if err != nil {
		return &RedisStorage{}, err
	}
	return &RedisStorage{
		db: client,
	}, nil
}

func (r RedisStorage) GetPremiumTime(userId int) (int, error) {
	result := r.db.Get(ctx, fmt.Sprintf("tg:premium:%v", userId))
	if result.Err() == redis.Nil {
		return -1, nil
	}
	if result.Err() != nil {
		return 0, result.Err()
	}
	time, err := result.Int()
	if err != nil {
		return 0, err
	}
	return time, nil
}

func (r RedisStorage) IncrementPremiumTime(userId, duration int) (int, error) {
	result := r.db.IncrBy(ctx, fmt.Sprintf("tg:premium:%v", userId), int64(duration))
	if result.Err() != nil {
		return 0, result.Err()
	}
	return int(result.Val()), nil
}

func (r RedisStorage) SavePremiumUser(userId int) error {
	status := r.db.Set(ctx, fmt.Sprintf("tg:premium:%v", userId), 0, time.Duration(monthDuration.Seconds()))
	return status.Err()
}

func (r RedisStorage) IsPremium(userId int) (int, error) {
	result := r.db.Get(ctx, fmt.Sprintf("tg:premium:%v", userId))
	if result.Err() == redis.Nil {
		return 0, nil
	}
	if result.Err() != nil {
		return 0, result.Err()
	}
	return 1, nil
}

func (r RedisStorage) GetUnpremiumTime(userId int) (int, error) {
	result := r.db.Get(ctx, fmt.Sprintf("tg:unpremium:time:%v", userId))
	// if user doesn't exists
	if result.Err() == redis.Nil {
		return -1, nil
	}
	if result.Err() != nil {
		return 0, result.Err()
	}
	time, err := result.Int()
	if err != nil {
		return 0, err
	}
	return time, nil
}

func (r *RedisStorage) SaveUnpremiumUser(userId int) error {
	err := r.db.Set(ctx, fmt.Sprintf("tg:unpremium:time:%v", userId), 0, monthDuration).Err()
	if err != nil {
		return err
	}
	err = r.db.Expire(ctx, fmt.Sprintf("tg:unpremium:time:%v", userId), monthDuration).Err()
	if err != nil {
		return err
	}
	return nil
}

func (r *RedisStorage) IncrementUnpremiumTime(userId, audioDuration int) (int, error) {
	time, err := r.db.IncrBy(ctx, fmt.Sprintf("tg:unpremium:time:%v", userId), int64(audioDuration)).Result()
	if err != nil {
		return 0, err
	}
	log.Printf("IncrementUnpremiumTime: %v", time)
	return int(time), nil
}
