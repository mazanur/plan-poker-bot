package bot_handler

import (
	"fmt"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"gotestbot/internal/bot/view"
	"gotestbot/internal/service"
	"gotestbot/internal/service/model"
	"gotestbot/sdk/tgbot"
	"strconv"
	"time"
)

type BotApp struct {
	view        *view.View
	roomService *service.RoomService
	taskService *service.TaskService
	rateService *service.RateService
}

func NewBotApp(view *view.View, roomProv *service.RoomService, taskProv *service.TaskService, rateProv *service.RateService) *BotApp {
	return &BotApp{view: view,
		roomService: roomProv,
		taskService: taskProv,
		rateService: rateProv,
	}
}

func (b *BotApp) Handle(u *tgbot.Update) {

	switch {
	case u.HasCommand("/start") || u.HasAction(view.ActionStart):
		_, _ = b.view.StartView(u)

	case u.HasActionOrChain(view.ActionCreateTask):
		b.HandleAddTask(u)

	case u.HasActionOrChain(view.ActionAddRate):
		roomId := u.GetButton().GetData("roomId")
		taskId := u.GetButton().GetData("taskId")
		sum := u.GetButton().GetData("sum")
		parse, _ := uuid.Parse(taskId)

		sumInt64, err := strconv.ParseInt(sum, 10, 32)
		if err != nil {
			log.Error().Err(err).Msg("")
		}
		rate := model.Rate{
			Id:          uuid.New(),
			UserId:      u.GetUserId(),
			TaskId:      parse,
			Sum:         int32(sumInt64),
			CreatedDate: time.Now(),
		}
		if err = b.rateService.UpsertRate(rate); err != nil {
			log.Error().Err(err).Msg("")
			b.sendErrorMessage(u)
			return
		}

		finished, err := b.taskService.TaskFinished(taskId)
		if err != nil {
			log.Error().Err(err).Msg("")
			return
		}

		if finished {
			err = b.taskService.SetFinished(taskId)
			if err != nil {
				log.Error().Err(err).Msg("")
				return
			}
			_, _ = b.view.ShowFinishedTaskView(taskId, roomId, u)
			_, _ = b.view.ShowSetTaskGrade(taskId, roomId, u)

		} else {
			//_, _ = b.view.ShowTaskTime(taskId, roomId, u)
			_, _ = b.view.ShowTaskView(0, taskId, roomId, u)
		}

	case u.HasChain(view.ActionRevoteTaskRate):
		roomId := u.GetButton().GetData("roomId")
		taskId := u.GetButton().GetData("taskId")

		err := b.rateService.DelRatesByTaskId(taskId)
		if err != nil {
			log.Error().Err(err).Msg("")
			_, _ = b.view.ErrorMessage(u, "Не получилось ресстартовать голосование")
			return
		}
		room, err := b.roomService.GetRoomById(roomId)
		if err != nil {
			log.Error().Err(err).Msgf("unable to get room by roomId: %d", roomId)
			return
		}

		_, err = b.view.ShowTaskView(room.ChatId, taskId, roomId, u)
		if err != nil {
			_, _ = b.view.ErrorMessage(u, "❗️ Не получилось опубликовать задачцу")
		} else {
			_, _ = b.view.ShowTaskView(room.ChatId, taskId, roomId, u)
			go b.view.ErrorMessage(u, "Задача успешно опубликована")
			_, _ = b.view.ShowRoomView("", roomId, u)
		}

	case u.HasAction(view.ActionShowRooms):
		_, _ = b.view.ShowRooms(u)

	case u.HasActionOrChain(view.ActionCreateRoom):
		b.HandleAddRoom(u)

	case u.HasAction(view.ActionShowRoom):
		roomId := u.GetButton().GetData("roomId")
		_, _ = b.view.ShowRoomView("", roomId, u)

	case u.HasAction(view.ActionShowTasks):
		roomId := u.GetButton().GetData("roomId")
		page, _ := strconv.Atoi(u.GetButton().GetData("page"))
		_, _ = b.view.ShowTasks(roomId, page, u)

	case u.HasAction(view.ActionJoinRoom):
		roomId := u.GetButton().GetData("roomId")
		if err := b.roomService.SaveRoomMember(u.GetUser().UserId, roomId); err != nil {
			log.Error().Err(err).Msg("")
			b.sendErrorMessage(u)
			return
		}
		_, _ = b.view.ShowRoomViewInline(roomId, u)

	case u.HasAction(view.ActionSetGroupOfRoom):
		roomId := u.GetButton().GetData("roomId")
		chatId := u.GetButton().GetData("chatId")
		chatIdInt64, _ := strconv.ParseInt(chatId, 10, 64)

		_, err := b.view.SendChatWritingAction(chatIdInt64)
		if err != nil {
			log.Error().Err(err).Msg("")
			_, _ = b.view.ErrorMessage(u, fmt.Sprintf("❗Сперва добавьте бота в чат %v", u.GetButton().GetData("chatName")))
			return
		}
		if err = b.roomService.SetChatIdRoom(roomId, chatIdInt64); err != nil {
			log.Error().Err(err).Msg("")
			b.sendErrorMessage(u)
			return
		}
		go b.view.ErrorMessage(u, fmt.Sprintf("✅ Чат %v успешно привязан", u.GetButton().GetData("chatName")))
		_, _ = b.view.ShowRoomView("", roomId, u)

	case u.HasAction(view.ActionFinishTask):
		roomId := u.GetButton().GetData("roomId")
		room, err := b.roomService.GetRoomById(roomId)
		if err != nil {
			log.Error().Err(err).Msgf("unable to get room by roomId: %d", roomId)
			return
		}
		if room.UserId != u.GetUserId() {
			log.Warn().Err(err).Msgf("not finished by not admin user: %d", u.GetUserId())
			_, _ = b.view.ErrorMessage(u, "❗️ Раскрыться может только администратор комнаты")
			return
		}

		taskId := u.GetButton().GetData("taskId")
		if err = b.taskService.SetFinished(taskId); err != nil {
			log.Error().Err(err).Msgf("unable to set finished for task: %d", taskId)
			return
		}

		err = b.taskService.SetFinished(taskId)
		if err != nil {
			log.Error().Err(err).Msg("")
			return
		}
		_, _ = b.view.ShowFinishedTaskView(taskId, roomId, u)
		_, _ = b.view.ShowSetTaskGrade(taskId, roomId, u)

	case u.HasAction(view.ActionFinishRoom):
		roomId := u.GetButton().GetData("roomId")
		room, err := b.roomService.GetRoomById(roomId)
		if err != nil {
			log.Error().Err(err).Msgf("unable to get room by roomId: %d", roomId)
			return
		}
		if room.Status == model.Finished {
			_, _ = b.view.ErrorMessage(u, "❗️ Планирование уже завершено")
			return
		}

		_, err = b.view.ShowTasksAfterFinishedRoom(roomId, u)
		if err == nil {
			if err = b.roomService.SetStatusRoom(model.Finished, roomId); err != nil {
				log.Error().Err(err).Msgf("unable to set finished for room: %d", roomId)
				return
			}
			_, _ = b.view.ErrorMessage(u, "Планирование успешно завершено")
		} else {
			_, _ = b.view.ErrorMessage(u, "Не удалось завершить планирование")
		}

	case u.HasActionOrChain(view.ActionFinishTaskRate):
		b.HandleAddTaskGrade(u)

	case u.HasAction(view.ActionNextTask):
		roomId := u.GetButton().GetData("roomId")
		task, err := b.taskService.GetNextNotFinishedTask(roomId)
		if err != nil {
			log.Error().Err(err).Msg("")
			_, _ = b.view.ErrorMessage(u, "❗️ Не найдено запланированных задач!")
			return
		}
		room, err := b.roomService.GetRoomById(roomId)
		if err != nil {
			log.Error().Err(err).Msgf("unable to get room by roomId: %d", roomId)
			return
		}
		if room.UserId != u.GetUserId() {
			_, _ = b.view.ErrorMessage(u, "")
			return
		}

		msg, err := b.view.ShowTaskView(room.ChatId, task.Id.String(), roomId, u)
		if err != nil {
			_, _ = b.view.ErrorMessage(u, "❗️ Не получилось опубликовать задачцу")
		} else {
			_, _ = b.view.ErrorMessage(u, "Задача успешно опубликована")
			chatIdForLink := strconv.FormatInt(room.ChatId, 10)[4:]
			messageLink := fmt.Sprintf("[Ссылка на задачу](https://t.me/c/%v/%v) \n\n", chatIdForLink, msg.MessageID)
			_, _ = b.view.ShowRoomView(messageLink, roomId, u)
		}
	}

	switch {
	case u.GetInline() != "":
		rooms, err := b.roomService.GetRoomsByNameAndUserId(u.GetInline(), u.GetUser().UserId)
		if err != nil {
			log.Error().Err(err).Msg("")
			_, _ = b.view.ErrorMessageText("Ошибка получения комнат.\n", u)
		}
		_, _ = b.view.ShowRoomsInline(rooms, u)

	}
}
