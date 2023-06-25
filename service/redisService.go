package service

import (
	"errors"
	"log"
	"tgbot/storage"
)

const (
	PremiumMaxTime   = 3600
	UnpremiumMaxTime = 300
)

var (
	ErrUserDoesntExist = errors.New("User doesn't exist")
	ErrMaxTimeExceeded = errors.New("Max time exceeded")
)

type RedisService struct {
	storage storage.RedisStorage
}

func NewRedisService() *RedisService {
	redisStorage, err := storage.NewStorage()
	if err != nil {
		log.Fatal(err)
	}
	return &RedisService{
		storage: *redisStorage,
	}
}

func (rs RedisService) IcrementPremiumTime(userId, duration int) (int, error) {
	time, err := rs.storage.IncrementPremiumTime(userId, duration)
	if err != nil {
		return 0, err
	}
	if time >= PremiumMaxTime {
		return time, ErrMaxTimeExceeded
	}
	return time, nil
}

func (rs RedisService) GetPremiumTime(userId int) (int, error) {
	time, err := rs.storage.GetPremiumTime(userId)
	if err != nil {
		return 0, err
	}
	return time, nil
}

func (rs RedisService) SaveUnpremiumUser(userId int) error {
	err := rs.storage.SaveUnpremiumUser(userId)
	if err != nil {
		return err
	}
	return nil
}

func (rs RedisService) SavePremiumUser(userId int) error {
	log.Println(userId)
	return rs.storage.SavePremiumUser(userId)
}

func (rs RedisService) IsPremium(userId int) (bool, error) {
	isPremium, err := rs.storage.IsPremium(userId)
	if err != nil {
		return false, err
	}
	return isPremium == 1, nil
}

func (rs RedisService) GetUnpremiumTimeSpent(userId int) (int, error) {
	timeSpent, err := rs.storage.GetUnpremiumTime(userId)
	if err != nil {
		return 0, err
	}
	if timeSpent == -1 {
		return 0, ErrUserDoesntExist
	}
	if timeSpent >= UnpremiumMaxTime {
		return 0, ErrMaxTimeExceeded
	}
	return timeSpent, nil
}

func (rs RedisService) IncrementUnpremiumTime(userId int, audioDuration int) (int, error) {
	time, err := rs.storage.IncrementUnpremiumTime(userId, audioDuration)
	if err != nil {
		return 0, err
	}
	if time >= UnpremiumMaxTime {
		return time, ErrMaxTimeExceeded
	}
	return time, nil
}
