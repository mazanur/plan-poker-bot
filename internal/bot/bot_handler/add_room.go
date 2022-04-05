package bot_handler

import (
	"fmt"
	"github.com/go-pkgz/lgr"
	"github.com/google/uuid"
	"gotestbot/internal/bot/view"
	"gotestbot/internal/service/model"
	"gotestbot/sdk/tgbot"
	"strconv"
	"time"
)

func (b *BotApp) HandleAddRoom(u *tgbot.Update) {

	if u.HasAction(view.ActionCreateRoom) {
		u.StartChain(string(view.ActionCreateRoom)).StartChainStep("NAME").FlushChatInfo()
		_, _ = b.view.AddRoomName(u)
		return
	}

	switch u.GetChainStep() {
	case "NAME":
		msg, _ := b.view.SetChatRoom(u)
		u.StartChainStep("SET_CHAT_FOR_ROOM").
			AddChainData("name", u.GetText()).
			AddChainData("messageId", strconv.Itoa(msg.MessageID)).
			FlushChatInfo()

	case "SET_CHAT_FOR_ROOM":
		if u.Update.Message != nil &&
			u.Update.Message.NewChatMembers != nil &&
			u.Update.Message.NewChatMembers[0].UserName == b.view.GetMe().UserName {

			u.AddChainData("chatId", strconv.FormatInt(u.GetChatId(), 10))
			_, _ = b.view.AddSettingRoom(fmt.Sprintf("Бот успешно привязан к чату - *%v*\n\n", u.Message.Chat.Title), u)

		} else if u.HasAction(view.ActionBotAdded) {
			_, _ = b.view.AddSettingRoom("", u)
		}
		u.StartChainStep("SETTING").FlushChatInfo()
		_, _ = b.view.NewDeleteMessage(u.GetChatId(), u.GetMessageId())

	case "SETTING":
		roomId := uuid.New()
		chatId64, _ := strconv.ParseInt(u.GetChainData("chatId"), 10, 64)
		if err := b.roomService.SaveRoom(model.Room{
			Id:          roomId,
			Name:        u.GetChainData("name"),
			UserId:      u.GetUser().UserId,
			ChatId:      chatId64,
			Status:      model.New,
			CreatedDate: time.Now(),
		}); err != nil {
			lgr.Printf("[ERROR] SaveRoom not")
			b.sendErrorMessage(u)
			return
		}

		if err1 := b.roomService.SaveRoomMember(u.GetUser().UserId, roomId.String()); err1 != nil {
			lgr.Printf("[ERROR] can not save room")
			b.sendErrorMessage(u)
			return
		}

		text := "Отлично, вы успешно создали комнату, теперь нажмите *Отправить в чат* и выберите вашу группу\n\n"
		msg, _ := b.view.ShowRoomView(text, roomId.String(), u)

		if chatId64 != 0 {
			u.FinishChain().FlushChatInfo()
		} else {
			u.StartChainStep("SEND_TO_CHAT").
				AddChainData("messageId", strconv.Itoa(msg.MessageID)).
				AddChainData("roomId", roomId.String()).
				FlushChatInfo()
		}

	case "SEND_TO_CHAT":
		if u.Update.Message != nil &&
			u.Update.Message.ViaBot != nil &&
			u.Update.Message.ReplyMarkup != nil &&
			u.Update.Message.ReplyMarkup.InlineKeyboard != nil &&
			u.Update.Message.ReplyMarkup.InlineKeyboard[0][0].Text == "Присоединиться" {

			buttonId := u.Update.Message.ReplyMarkup.InlineKeyboard[0][0].CallbackData
			room, err := b.roomService.GetRoomById(u.GetButtonById(*buttonId).GetData("roomId"))
			if err != nil {
				lgr.Printf("[ERROR] ")
				b.sendErrorMessage(u)
				return
			}

			//chatId not editing
			if room.ChatId == u.Update.Message.Chat.ID {
				return
			}
			messageId, _ := strconv.Atoi(u.GetChainData("messageId"))
			_, _ = b.view.NewDeleteMessage(room.UserId, messageId)
			_, _ = b.view.ChangeChatOfRoom(room, u.Update.Message.Chat, u)

			u.FinishChain().FlushChatInfo()
		}
	}
}

func (b *BotApp) sendErrorMessage(u *tgbot.Update) {
	if u.IsButton() {
		_, _ = b.view.ErrorMessage(u, "Не удалось сохранить комнату\n")
	} else {
		_, _ = b.view.ErrorMessageText("Не удалось сохранить комнату\n", u)
	}
	return
}
