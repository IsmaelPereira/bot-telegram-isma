package clients

import (
	"log"
	"strings"

	"github.com/ismaelpereira/telegram-bot-isma/api/common"
)

func NewReminder(
	chatID string,
	reminderTitle string,
) error {
	redis, err := common.SetRedis()
	if err != nil {
		return err
	}
	key := "checklist:" + chatID + ":" + strings.TrimSpace(reminderTitle)
	if err = redis.Set(key, nil, 0).Err(); err != nil {
		return err
	}
	return nil
}

func AddReminder(
	chatID string,
	reminderTitle string,
	values []byte,
) error {
	redis, err := common.SetRedis()
	if err != nil {
		return err
	}
	key := "checklist:" + chatID + ":" + strings.TrimSpace(reminderTitle)
	if err = redis.Set(key, values, 0).Err(); err != nil {
		return err
	}
	return nil
}

func ListReminder(
	chatID string,
) ([]string, error) {
	redis, err := common.SetRedis()
	if err != nil {
		return nil, err
	}
	key := "checklist:" + chatID + ":*"
	keys, err := redis.Keys(key).Result()
	if err != nil {
		return nil, err
	}
	return keys, nil
}

func DeleteReminder(
	chatID string,
	reminderTitle string,
) error {
	redis, err := common.SetRedis()
	if err != nil {
		return err
	}
	key := "checklist:" + chatID + ":" + strings.TrimSpace(reminderTitle)
	if err = redis.Del(key).Err(); err != nil {
		return err
	}
	return nil
}

func GetReminder(reminderTitle string) ([]byte, error) {
	redis, err := common.SetRedis()
	if err != nil {
		return nil, err
	}
	list, err := redis.Get(reminderTitle).Bytes()
	if err != nil {
		log.Println(err)
		return nil, err
	}
	return list, nil
}
