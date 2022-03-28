package tgbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/rs/zerolog/log"
	"strings"
)

type Update struct {
	tgbotapi.Update
	chatProv ChatProvider

	chat *ChatInfo
	btn  *Button
	usr  User
}

func WrapUpdate(update tgbotapi.Update, user User, chatProvider ChatProvider) *Update {
	return &Update{Update: update, usr: user, chatProv: chatProvider}
}

func (u *Update) GetUserId() int64 {
	if u.Message != nil && u.Message.Chat != nil {
		return u.Message.From.ID
	}
	if u.CallbackQuery != nil {
		return u.CallbackQuery.From.ID
	}
	if u.InlineQuery != nil {
		return u.InlineQuery.From.ID
	}
	return 0
}

func (u *Update) GetChatId() int64 {
	if u.Message != nil && u.Message.Chat != nil {
		return u.Message.Chat.ID
	}
	if u.CallbackQuery != nil {
		return u.CallbackQuery.Message.Chat.ID
	}
	return 0
}

func (u *Update) GetUser() User {
	return u.usr
}

func (u *Update) GetMessageId() int {
	if u.IsButton() && u.CallbackQuery != nil {
		return u.CallbackQuery.Message.MessageID
	} else if u.Message != nil {
		return u.Message.MessageID
	}
	return 0
}

func (u *Update) GetInlineMessageId() string {
	if u.InlineQuery != nil {
		return u.InlineQuery.ID
	}
	return ""
}

func (u *Update) HasText(text string) bool {
	return text == u.Update.Message.Text
}

func (u *Update) HasCommand(text string) bool {
	return u.IsCommand() && text == u.Update.Message.Text
}

func (u *Update) IsCommand() bool {
	return u.Update.Message != nil &&
		strings.Contains(u.Update.Message.Text, "/")
}

//Button

func (u *Update) IsPlainText() bool {
	return !u.IsCommand() && u.Update.Message != nil && u.Update.Message.Text != ""
}

func (u *Update) GetText() string {
	if u.Message == nil {
		return ""
	}
	return u.Message.Text
}

func (u *Update) GetInline() string {
	if u.InlineQuery != nil {
		return u.InlineQuery.Query
	}
	return ""
}

func (u *Update) GetInlineId() string {
	if u.CallbackQuery != nil {
		return u.CallbackQuery.InlineMessageID
	}
	if u.InlineQuery != nil {
		return u.InlineQuery.ID
	}
	return ""
}

func (u *Update) IsButton() bool {
	return u.Update.CallbackData() != ""
}

func (u *Update) GetButton() Button {
	if u.btn == nil {
		button, err := u.chatProv.GetButton(u.CallbackData())
		if err != nil {
			log.Error().Err(err).Msgf("cannot find button %s", u.CallbackData())
		}
		u.btn = &button
	}
	return *u.btn
}

func (u *Update) GetButtonById(btnId string) Button {
	button, err := u.chatProv.GetButton(btnId)
	if err != nil {
		log.Error().Err(err).Msgf("cannot find button %s", u.CallbackData())
	}

	return button
}

func (u *Update) HasAction(action Action) bool {
	return u.IsButton() && u.GetButton().HasAction(action)
}

// ChatInfo

func (u *Update) HasActionOrChain(actionOrChain Action) bool {
	return u.IsButton() && u.GetButton().HasAction(actionOrChain) ||
		u.GetChatInfo().ActiveChain == string(actionOrChain)
}

func (u *Update) HasChain(chain Action) bool {
	return u.GetChatInfo().ActiveChain == string(chain)
}

func (u *Update) GetChatInfo() *ChatInfo {
	if u.chat == nil {
		chat, err := u.chatProv.GetChat(u.GetUserId())
		if err != nil {
			log.Error().Err(err).Msgf("cannot find chat chat")
		}
		u.chat = &chat
	}

	if u.chat.ChatId == 0 {
		u.chat.ChatId = u.GetUserId()
	}
	if u.chat.ChainData == nil {
		u.chat.ChainData = Data{}
	}

	return u.chat
}

func (u *Update) FlushChatInfo() {
	err := u.chatProv.SaveChatInfo(*u.GetChatInfo())
	if err != nil {
		log.Error().Err(err).Msgf("cannot save chat info: %+v", u.GetChatInfo())
	}
}

func (u *Update) StartChain(chain string) *Update {
	u.GetChatInfo().ActiveChain = chain
	return u
}

func (u *Update) StartChainStep(chainStep string) *Update {
	u.GetChatInfo().ActiveChainStep = chainStep
	return u
}

func (u *Update) GetChainStep() string {
	return u.GetChatInfo().ActiveChainStep
}

func (u *Update) AddChainData(key string, value string) *Update {
	u.GetChatInfo().ChainData[key] = value
	return u
}

func (u *Update) GetChainData(key string) string {
	return u.GetChatInfo().ChainData[key]
}

func (u *Update) FinishChain() *Update {
	u.GetChatInfo().ActiveChain = ""
	u.GetChatInfo().ActiveChainStep = ""
	u.GetChatInfo().ChainData = map[string]string{}
	return u
}
