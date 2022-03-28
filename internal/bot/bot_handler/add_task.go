package bot_handler

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gotestbot/internal/bot/view"
	"gotestbot/internal/service/model"
	"gotestbot/sdk/tgbot"
	"strconv"
	"time"
)

func (b *BotApp) HandleAddTask(u *tgbot.Update) {

	if u.HasAction(view.ActionCreateTask) {
		roomId := u.GetButton().GetData("roomId")
		room, err := b.roomService.GetRoomById(roomId)
		if err != nil {
			log.Error().Err(err).Msgf("unable to get room by roomId: %d", roomId)
			return
		}
		if room.ChatId == 0 {
			_, _ = b.view.ErrorMessage(u, "❗️ Перед тем как добавить задачу, необходимо 'Отправить в чат', а также добавить бота в чат")
			return
		}
		if room.Status == model.Finished {
			_, _ = b.view.ErrorMessage(u, "❗️ Планирование уже завершено")
			return
		}

		u.StartChain(string(view.ActionCreateTask)).
			StartChainStep("NAME").AddChainData("roomId", roomId).FlushChatInfo()
		_, _ = b.view.AddTaskName(u)
		return
	}

	switch u.GetChainStep() {
	case "NAME":
		u.StartChainStep("URL").AddChainData("name", u.GetText()).FlushChatInfo()
		_, _ = b.view.AddTaskUrl(u)

	case "URL":
		u.StartChainStep("SETTING").AddChainData("url", u.GetText()).FlushChatInfo()
		_, _ = b.view.AddSettingTask("", u)

	case "SETTING":
		if !u.HasAction(view.ActionSaveAndSendTask) &&
			!u.HasAction(view.ActionSaveAndSaveTask) &&
			!u.HasAction(view.ActionSaveTaskAndCancel) {
			return
		}

		takId := uuid.New()
		roomId := u.GetChainData("roomId")
		roomIdUuid, _ := uuid.Parse(roomId)
		if err := b.taskService.SaveTask(model.Task{
			Id:          takId,
			Name:        u.GetChainData("name"),
			Url:         u.GetChainData("url"),
			RoomId:      roomIdUuid,
			CreatedDate: time.Now(),
		}); err != nil {
			log.Error().Err(err).Msg("")
			b.sendErrorMessage(u)
			return
		}

		room, err := b.roomService.GetRoomById(roomId)
		if err != nil {
			log.Error().Err(err).Msgf("unable to get room by roomId: %d", roomId)
			return
		}

		switch {
		case u.HasAction(view.ActionSaveAndSendTask):
			msg, err := b.view.ShowTaskView(room.ChatId, takId.String(), roomId, u)
			if err != nil {
				_, _ = b.view.ErrorMessage(u, "Не получилось опубликовать задачу")
			} else {
				u.FinishChain().FlushChatInfo()
				_, _ = b.view.ErrorMessage(u, "Задача успешно опубликована")
				chatIdForLink := strconv.FormatInt(room.ChatId, 10)[4:]
				messageLink := fmt.Sprintf("[Ссылка на задачу](https://t.me/c/%v/%v) \n\n", chatIdForLink, msg.MessageID)
				_, _ = b.view.ShowRoomView(messageLink, roomId, u)
			}
		case u.HasAction(view.ActionSaveAndSaveTask):
			_, _ = b.view.ErrorMessage(u, "Задача успешно сохранена")

			u.FinishChain().FlushChatInfo()
			u.StartChain(string(view.ActionCreateTask)).StartChainStep("NAME").
				AddChainData("roomId", roomId).FlushChatInfo()
			_, _ = b.view.AddTaskName(u)

		default:
			u.FinishChain().FlushChatInfo()
			_, _ = b.view.ErrorMessage(u, "Задача успешно сохранена")
			_, _ = b.view.ShowRoomView("", roomId, u)
		}

	}

}
