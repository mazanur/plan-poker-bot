package tgbot

import (
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type MessageBuilder struct {
	editMessage bool
	chatId      int64
	messageId   int
	inlineId    string
	text        string
	keyboard    [][]tgbotapi.InlineKeyboardButton
}

func (b *MessageBuilder) EditMessageTextAndMarkup(chatId int64, messageId int) *MessageBuilder {
	b.chatId = chatId
	b.messageId = messageId
	b.editMessage = true
	return b
}

func (b *MessageBuilder) NewMessage(chatId int64) *MessageBuilder {
	b.chatId = chatId
	b.editMessage = false
	return b
}

func (b *MessageBuilder) Message(chatId int64, messageId int) *MessageBuilder {
	if messageId == 0 {
		return b.NewMessage(chatId)
	} else {
		return b.EditMessageTextAndMarkup(chatId, messageId)
	}
}

func (b *MessageBuilder) Text(text string) *MessageBuilder {
	b.text = text
	return b
}

func (b *MessageBuilder) ChatId(chatId int64) *MessageBuilder {
	b.chatId = chatId
	return b
}

func (b *MessageBuilder) MessageId(messageId int) *MessageBuilder {
	b.messageId = messageId
	return b
}

func (b *MessageBuilder) InlineId(inlineId string) *MessageBuilder {
	b.inlineId = inlineId
	return b
}

func (b *MessageBuilder) Edit(editMessage bool) *MessageBuilder {
	b.editMessage = editMessage
	return b
}

func (b *MessageBuilder) AddKeyboardRow() *MessageBuilder {
	b.keyboard = append(b.keyboard, []tgbotapi.InlineKeyboardButton{})
	return b
}

func (b *MessageBuilder) AddKeyboard(keyboard [][]tgbotapi.InlineKeyboardButton) *MessageBuilder {
	b.keyboard = keyboard
	return b
}

func (b *MessageBuilder) AddButton(text, callbackData string) *MessageBuilder {
	b.keyboard[len(b.keyboard)-1] = append(b.keyboard[len(b.keyboard)-1],
		tgbotapi.InlineKeyboardButton{Text: text, CallbackData: &callbackData})
	return b
}

func (b *MessageBuilder) AddButtonUrl(text, url string) *MessageBuilder {
	b.keyboard[len(b.keyboard)-1] = append(b.keyboard[len(b.keyboard)-1],
		tgbotapi.InlineKeyboardButton{Text: text, URL: &url})
	return b
}

func (b *MessageBuilder) AddButtonSwitch(text, sw string) *MessageBuilder {
	b.keyboard[len(b.keyboard)-1] = append(b.keyboard[len(b.keyboard)-1],
		tgbotapi.NewInlineKeyboardButtonSwitch(text, sw),
	)
	return b
}

func (b *MessageBuilder) Build() tgbotapi.Chattable {
	if b.editMessage {
		kb := b.getKeyboard()
		var msg tgbotapi.Chattable
		if len(kb) > 0 {
			m := tgbotapi.NewEditMessageTextAndMarkup(
				b.chatId, b.messageId, b.text, tgbotapi.NewInlineKeyboardMarkup(kb...))
			m.ParseMode = tgbotapi.ModeMarkdown
			m.InlineMessageID = b.inlineId
			msg = m
		} else {
			m := tgbotapi.NewEditMessageText(b.chatId, b.messageId, b.text)
			m.ParseMode = tgbotapi.ModeMarkdown
			m.InlineMessageID = b.inlineId
			msg = m
		}

		return msg
	} else {
		msg := tgbotapi.NewMessage(b.chatId, b.text)
		keyboard := b.getKeyboard()
		if len(keyboard) > 0 {
			msg.ReplyMarkup = tgbotapi.NewInlineKeyboardMarkup(keyboard...)
		}
		msg.ParseMode = tgbotapi.ModeMarkdown
		return msg
	}
}

func (b *MessageBuilder) getKeyboard() [][]tgbotapi.InlineKeyboardButton {
	var keyboard [][]tgbotapi.InlineKeyboardButton

	for _, buttons := range b.keyboard {
		if len(buttons) > 0 {
			keyboard = append(keyboard, buttons)
		}
	}
	return keyboard
}

type inlineMessageBuilder struct {
	inlineQueryId string
	articles      []*tgbotapi.InlineQueryResultArticle
}

func NewInlineRequest(inlineQueryId string) *inlineMessageBuilder {
	return &inlineMessageBuilder{inlineQueryId: inlineQueryId}
}

func (b *inlineMessageBuilder) AddArticle(id, title, descr, text string) *inlineMessageBuilder {
	article := tgbotapi.NewInlineQueryResultArticleMarkdown(id, title, text)
	article.Description = descr
	b.articles = append(b.articles, &article)
	return b
}

func (b *inlineMessageBuilder) getLastArticleMarkup() *tgbotapi.InlineKeyboardMarkup {
	article := b.articles[len(b.articles)-1]
	if article.ReplyMarkup != nil {
		return article.ReplyMarkup
	} else {
		markup := tgbotapi.NewInlineKeyboardMarkup()
		article.ReplyMarkup = &markup
		return article.ReplyMarkup
	}
}

func (b *inlineMessageBuilder) AddKeyboardRow() *inlineMessageBuilder {
	markup := b.getLastArticleMarkup()
	markup.InlineKeyboard = append(markup.InlineKeyboard, []tgbotapi.InlineKeyboardButton{})
	return b
}

func (b *inlineMessageBuilder) AddButton(text, callbackData string) *inlineMessageBuilder {

	markup := b.getLastArticleMarkup()

	markup.InlineKeyboard[len(markup.InlineKeyboard)-1] = append(markup.InlineKeyboard[len(markup.InlineKeyboard)-1],
		tgbotapi.InlineKeyboardButton{Text: text, CallbackData: &callbackData})
	return b
}

func (b *inlineMessageBuilder) Build() tgbotapi.Chattable {

	var articles []interface{}
	for _, article := range b.articles {
		articles = append(articles, *article)
	}

	return tgbotapi.InlineConfig{
		InlineQueryID: b.inlineQueryId,
		IsPersonal:    true,
		Results:       articles,
	}

}
