package controllers

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"github.com/go-redis/redis/v7"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api"
	"github.com/ismaelpereira/telegram-bot-isma/api/clients"
	"github.com/ismaelpereira/telegram-bot-isma/bot/msgs"
	"github.com/ismaelpereira/telegram-bot-isma/config"
	"github.com/ismaelpereira/telegram-bot-isma/types"
)

var animes []types.Anime

// AnimeHandleUpdate is a function for anime work
func AnimesHandleUpdate(
	cfg *config.Config,
	redis *redis.Client,
	bot *tgbotapi.BotAPI,
	update *tgbotapi.Update,
) error {
	if update.CallbackQuery != nil {
		return animesArrowButtonAction(cfg, update, animes)
	}
	command := update.Message.Command()
	animeName := strings.TrimSpace(update.Message.CommandArguments())
	if animeName == "" {
		msg := tgbotapi.NewMessage(update.Message.Chat.ID, msgs.MsgAnimes)
		_, err := bot.Send(msg)
		return err
	}
	jikanAPI, err := clients.NewJikanAPI(animeName, command)
	if err != nil {
		return err
	}
	_, animes, err = jikanAPI.SearchAnimeOrManga(animeName, command)
	if err != nil {
		return err
	}
	if len(animes) == 0 {
		return nil
	}
	animeMessage := getAnimesPictureAndSendMessage(update, animes[0])
	var kb []tgbotapi.InlineKeyboardMarkup
	if len(animes) > 1 {
		kb = append(kb, tgbotapi.NewInlineKeyboardMarkup(
			tgbotapi.NewInlineKeyboardRow(
				tgbotapi.NewInlineKeyboardButtonData(msgs.IconNext, "animes:1"),
			),
		))
	}
	if len(animes) > 1 {
		animeMessage.ReplyMarkup = kb[0]
	}
	_, err = bot.Send(animeMessage)
	return err
}

func getAnimesPictureAndSendMessage(
	update *tgbotapi.Update,
	an types.Anime,
) *tgbotapi.PhotoConfig {
	var anMessage tgbotapi.PhotoConfig
	if update.CallbackQuery == nil {
		anMessage = tgbotapi.NewPhotoShare(update.Message.Chat.ID, an.CoverPicture)
	}
	if update.CallbackQuery != nil {
		anMessage = tgbotapi.NewPhotoShare(update.CallbackQuery.Message.Chat.ID, an.CoverPicture)
	}
	var airing string
	if an.Airing {
		airing = "Passando"
	} else {
		airing = "Finalizado"
	}
	animeEpisodes := strconv.Itoa(an.Episodes)
	if animeEpisodes == "0" {
		animeEpisodes = "?"
	}
	anMessage.Caption = "Título: " + an.Title +
		"\nNota: " + strconv.FormatFloat(an.Score, 'f', 2, 64) +
		"\nEpisódios: " + animeEpisodes +
		"\nStatus: " + airing
	return &anMessage
}

func animesArrowButtonAction(
	cfg *config.Config,
	update *tgbotapi.Update,
	animes []types.Anime,
) error {
	i, err := strconv.Atoi(update.CallbackQuery.Data)
	if err != nil {
		return err
	}
	animeMessage := getAnimesPictureAndSendMessage(update, animes[i])
	var kb []tgbotapi.InlineKeyboardButton
	if i != 0 {
		kb = append(kb,
			tgbotapi.NewInlineKeyboardButtonData(msgs.IconPrevious, "animes:"+strconv.Itoa(i-1)),
		)
	}
	if i != (len(animes) - 1) {
		kb = append(kb,
			tgbotapi.NewInlineKeyboardButtonData(msgs.IconNext, "animes:"+strconv.Itoa(i+1)),
		)
	}
	var msgEdit types.EditMediaJSON
	msgEdit.ChatID = update.CallbackQuery.Message.Chat.ID
	msgEdit.MessageID = update.CallbackQuery.Message.MessageID
	msgEdit.Media.Type = "photo"
	if animes[i].CoverPicture == "" {
		msgEdit.Media.URL = "https://badybassitt.sp.gov.br/lib/img/no-image.jpg"
	} else {
		msgEdit.Media.URL = animes[i].CoverPicture
	}
	msgEdit.Media.Caption = animeMessage.Caption
	msgEdit.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(
		kb,
	)
	messageJSON, err := json.Marshal(msgEdit)
	if err != nil {
		return err
	}
	sendMessage, err := http.Post("https://api.telegram.org/bot"+cfg.Telegram.Key+"/editMessageMedia",
		"application/json", bytes.NewBuffer(messageJSON))
	if err != nil {
		return err
	}
	defer sendMessage.Body.Close()
	if sendMessage.StatusCode > 299 || sendMessage.StatusCode < 200 {
		err = fmt.Errorf("Error in post method %w", err)
		log.Println(err)
		return err
	}
	return nil
}
