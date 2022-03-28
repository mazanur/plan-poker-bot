package view

import (
	"encoding/json"
	"fmt"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gotestbot/sdk/tgbot"
)

func (v *View) createButton(action tgbot.Action, data map[string]string) *tgbot.Button {
	id := uuid.New()
	button := tgbot.Button{
		Id:     id.String(),
		Action: action,
		Data:   data,
	}
	err := v.chatProv.SaveButton(button)
	if err != nil {
		log.Error().Err(err).Msgf("cannot save button")
	}
	return &button
}

func logIfError(send tgbotapi.Message, err error) (tgbotapi.Message, error) {
	if err == nil {
		return send, nil
	}
	switch err.(type) {
	default:
		log.Error().Err(err).Stack().Msg("cannot send")
		return send, err

	case *json.UnmarshalTypeError:
		log.Warn().Err(err).Stack().Msg("")
		return send, nil
	}
}

func userLink(user *tgbot.User) string {
	return fmt.Sprintf("[%s](tg://user?id=%d)", user.DisplayName, user.UserId)
}
