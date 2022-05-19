package view

import (
	"fmt"
	"github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gotestbot/sdk/tgbot"
)

func (v *View) AddRoomName(u *tgbot.Update) (tgbotapi.Message, error) {
	builder := new(tgbot.MessageBuilder).
		Message(u.GetUserId(), u.GetMessageId()).
		Text("Введите название комнаты")

	return logIfError(v.tg.Send(builder.Build()))
}

func (v *View) AddSettingRoom(prefix string, u *tgbot.Update) (tgbotapi.Message, error) {
	timerBtn := v.createButton(ActionRoomSettingTimes, nil)

	builder := new(tgbot.MessageBuilder).
		NewMessage(u.GetUserId()).
		Text(prefix+"Выберите настройка для комнаты").
		AddKeyboardRow().AddButton("⏳ Использовать таймер", timerBtn.Id).
		AddKeyboardRow().AddButton("❌ Не использовать таймер", timerBtn.Id)

	return logIfError(v.tg.Send(builder.Build()))
}

func (v *View) SetChatRoom(u *tgbot.Update) (tgbotapi.Message, error) {
	timerBtn := v.createButton(ActionBotAdded, nil)

	builder := new(tgbot.MessageBuilder).
		NewMessage(u.GetUserId()).
		Text("Добавьте бота в группу, в которой хотите проводить планирование").
		AddKeyboardRow().AddButton("Бот был добавлен ранее", timerBtn.Id)

	return logIfError(v.tg.Send(builder.Build()))
}

func (v *View) ShowRooms(u *tgbot.Update) (tgbotapi.Message, error) {
	roomId := "roomId"
	users, err := v.roomProv.GetUsersByRoomId(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to get users by roomId: %d, $v", roomId, err)
		return tgbotapi.Message{}, err
	}

	var members string
	for _, user := range users {
		members += "- " + userLink(&user) + "\n"
	}
	room, err := v.roomProv.GetRoomById(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to get room by roomId: %d, $v", roomId, err)
		return tgbotapi.Message{}, err
	}

	builder := new(tgbot.MessageBuilder).
		InlineId(u.GetInlineId()).
		Edit(u.IsButton()).
		Text(fmt.Sprintf("Комната - *%v*\n🗓 %v \n\nУчастники:\n%v", room.Name, room.CreatedDate.Format("02 January 2006"), members))

	joinBtn := v.createButton(ActionJoinRoom, map[string]string{"roomId": room.Id.String()})

	builder.AddKeyboardRow().AddButton("Присоединиться", joinBtn.Id).Build()
	send, err := v.tg.Send(builder.Build())
	return logIfError(send, err)
}
