package utils

import (
	"log"
	"time"

	"github.com/go-redis/redis"
)

// SaveDataToRedis : receives sessionID, json string to save and redisClient to save to redis. Returns error if sessionID already exists
func SaveDataToRedis(uniqueSessionID string, UssdData string, client *redis.Client, expire_seconds time.Duration) error {
	err := client.Set(uniqueSessionID, UssdData, 0).Err()
	log.Printf("[r] Saving session to redis : %v", UssdData)
	if err != nil {
		log.Println("error setting key", uniqueSessionID, "==>", err)
		return err
	}
	_ = client.Expire(uniqueSessionID, expire_seconds*time.Second)
	log.Println("Data addded to redis")
	return nil
}

// FetchDataFromRedis : returns a key value from redis
func FetchDataFromRedis(uniqueSessionID string, client *redis.Client) (string, error) {
	val, err := client.Get(uniqueSessionID).Result()
	if err != nil {
		log.Println("error fetching key", uniqueSessionID, "==>", err)
		return "", err
	}

	return val, nil
}

//Delete session from redis
func RemoveDataFromRedis(uniqueSessionID string, client *redis.Client) error {
	err := client.Del(uniqueSessionID).Err()
	if err != nil {
		log.Println("error deleting key", uniqueSessionID, "==>", err)
		return err
	}
	return nil
}
