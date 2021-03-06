package controllers

import (
	"log"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-redis/redis/v7"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ismaelpereira/telegram-bot-isma/bot/msgs"
	"github.com/ismaelpereira/telegram-bot-isma/config"
)

func TimerHandleUpdate(
	cfg *config.Config,
	redis *redis.Client,
	bot *tgbotapi.BotAPI,
	update *tgbotapi.Update,
) error {
	if update.Message.Command() == "reminder" {
		if update.Message.CommandArguments() == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgs.MsgReminder)
			_, err := bot.Send(msg)
			return err
		}
		return reminderHandler(redis, bot, update)
	}
	if update.Message.Command() == "now" {
		if strings.ToLower(update.Message.CommandArguments()) == "" {
			msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgs.MsgNow)
			_, err := bot.Send(msg)
			return err
		}
		return nowHandler(bot, update)
	}
	return nil
}

func nowHandler(bot *tgbotapi.BotAPI, update *tgbotapi.Update) error {
	commandSplited := strings.Fields(strings.ToLower(update.Message.CommandArguments()))
	if len(commandSplited) != 3 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Você digitou o comando errado. Não foi possível completar a solicitação")
		_, err := bot.Send(msg)
		return err
	}
	operation := commandSplited[0]
	value := commandSplited[1]
	measureOfTime := commandSplited[2]
	var hour time.Time
	switch operation {
	case "plus":
		{
			duration, err := time.ParseDuration(value + string(measureOfTime[0]))
			if err != nil {
				return err
			}
			hour = time.Now().Add(duration).Add(-time.Hour * 3)
		}
	case "minus":
		{
			duration, err := time.ParseDuration(value + string(measureOfTime[0]))
			if err != nil {
				return err
			}
			hour = time.Now().Add(-duration).Add(-time.Hour * 3)
		}
	default:
		{
			msg := tgbotapi.NewMessage(update.Message.Chat.ID,
				"Você digitou o comando errado. Não foi possível completar a solicitação")
			_, err := bot.Send(msg)
			return err
		}
	}

	time := hour.Format("Monday, 2 January, 2006 - 15:04:05")
	msg := tgbotapi.NewMessage(update.Message.Chat.ID, time)
	_, err := bot.Send(msg)
	return err
}

func reminderHandler(
	redis *redis.Client,
	bot *tgbotapi.BotAPI,
	update *tgbotapi.Update,
) error {
	commandSplited := strings.SplitAfterN(strings.ToLower(update.Message.CommandArguments()), " ", 3)
	if len(commandSplited) < 3 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Você digitou o comando errado. Não foi possível completar a solicitação")
		_, err := bot.Send(msg)
		return err
	}
	value := strings.TrimSpace(commandSplited[0])
	t, err := strconv.Atoi(value)
	if err != nil {
		return err
	}
	if t < 0 {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID,
			"Não é possível criar o lembrete com o tempo negativo")
		_, err := bot.Send(msg)
		return err
	}
	measureOfTime := strings.TrimSpace(commandSplited[1])
	message := commandSplited[2]
	var expireTime time.Time
	duration, err := time.ParseDuration(value + string(measureOfTime[0]))
	if err != nil {
		return err
	}
	expireTime = time.Now().Add(duration)
	if err := redis.HMSet("telegram:reminder:"+
		expireTime.Format("2006:01:02:15:04:05"), "chatID", update.Message.Chat.ID,
		"text", message).Err(); err != nil {
		return err
	}
	msg := tgbotapi.NewMessage(update.Message.Chat.ID,
		"Lembrete criado com sucesso! \nPara: "+expireTime.Format("02/01/2006 - 15:04:05")+"\nCom o texto: "+message)
	_, err = bot.Send(msg)
	return err
}

func reminderWorker(bot *tgbotapi.BotAPI, redis *redis.Client) {
	keys, err := redis.Keys("telegram:reminder:*").Result()
	if err != nil {
		return
	}
	sort.Strings(keys)
	now := "telegram:reminder:" + time.Now().Format("2006:01:02:15:04:05")
	for _, key := range keys {
		if key > now {
			continue
		}
		log.Printf("got reminder with key %q\n", key)
		data, err := redis.HGetAll(key).Result()
		if err != nil {
			return
		}
		if data != nil && data["chatID"] != "" {
			chatID, err := strconv.ParseInt(data["chatID"], 10, 64)
			if err != nil {
				return
			}
			msg := tgbotapi.NewMessage(chatID, msgs.IconAlarmClock+data["text"])
			if _, err = bot.Send(msg); err != nil {
				return
			}
			if err = redis.Del(key).Err(); err != nil {
				return
			}
		}
	}
}

func ReminderCheck(bot *tgbotapi.BotAPI, redis *redis.Client) {
	for {
		reminderWorker(bot, redis)
	}
}
