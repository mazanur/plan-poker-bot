package bot_handler

import (
	"fmt"
	log "github.com/go-pkgz/lgr"
	"github.com/google/uuid"
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

func (b *BotApp) Handle(u *tgbot.Update) error {

	switch {
	case u.HasCommand("/start") || u.HasAction(view.ActionStart):
		u.FinishChain().FlushChatInfo()
		_, _ = b.view.StartView(u)

	case u.HasActionOrChain(view.ActionCreateTask):
		b.HandleAddTask(u)

	case u.HasActionOrChain(view.ActionAddRate):
		if err := b.handleAddRate(u); err != nil {
			return err
		}

	case u.HasAction(view.ActionRevoteTaskRate):
		if err := b.handleRevoteTaskRate(u); err != nil {
			return err
		}

	case u.HasAction(view.ActionShowRooms):
		_, _ = b.view.ShowRooms(u)

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
			log.Printf("[ERROR] %v", err)
			return err
		}
		_, _ = b.view.ShowRoomViewInline(roomId, u)

	case u.HasAction(view.ActionSetGroupOfRoom):
		roomId := u.GetButton().GetData("roomId")
		chatId := u.GetButton().GetData("chatId")
		chatIdInt64, _ := strconv.ParseInt(chatId, 10, 64)

		_, err := b.view.SendChatWritingAction(chatIdInt64)
		if err != nil {
			log.Printf("[ERROR]  %v", err)
			_, _ = b.view.ErrorMessage(u, fmt.Sprintf("❗Сперва добавьте бота в чат %v", u.GetButton().GetData("chatName")))
			return err
		}
		if err = b.roomService.SetChatIdRoom(roomId, chatIdInt64); err != nil {
			log.Printf("[ERROR]  %v", err)
			return err
		}
		go b.view.ErrorMessage(u, fmt.Sprintf("✅ Чат %v успешно привязан", u.GetButton().GetData("chatName")))
		_, _ = b.view.ShowRoomView("", roomId, u)

	case u.HasAction(view.ActionFinishTask):
		roomId := u.GetButton().GetData("roomId")
		room, err := b.roomService.GetRoomById(roomId)
		if err != nil {
			return err
		}
		if room.UserId != u.GetUserId() {
			log.Printf("[WARN] not finished by not admin user: %d, %v", u.GetUserId(), err)
			_, _ = b.view.ErrorMessage(u, "❗️ Раскрыться может только администратор комнаты")
			return nil
		}

		taskId := u.GetButton().GetData("taskId")
		if err = b.taskService.SetFinished(taskId); err != nil {
			log.Printf("[ERROR] unable to set room by roomId: %d, %v", roomId, err)
			return err
		}

		rates, err := b.rateService.GetRatesByTaskId(taskId)
		if err != nil {
			return err
		}
		if rates == nil {
			_, _ = b.view.ErrorMessage(u, "❗️ Невозможно завершить оценку задачи, отсутствуют оценки")
			return nil
		}

		err = b.taskService.SetFinished(taskId)
		if err != nil {
			log.Printf("[ERROR]  %v", err)
			return err
		}
		_, _ = b.view.ShowFinishedTaskView(taskId, roomId, rates, u)
		_, _ = b.view.ShowSetTaskGrade(taskId, roomId, u)

	case u.HasAction(view.ActionFinishRoom):
		if err := b.handleFinishRoom(u); err != nil {
			return err
		}

	case u.HasActionOrChain(view.ActionFinishTaskRate):
		b.HandleAddTaskGrade(u)

	case u.HasAction(view.ActionNextTask):
		if err := b.handleNextTask(u); err != nil {
			return err
		}
	}

	switch {
	case u.GetInline() != "":
		rooms, err := b.roomService.GetRoomsByNameAndUserId(u.GetInline(), u.GetUser().UserId)
		if err != nil {
			log.Printf("[ERROR]  %v", err)
			_, _ = b.view.ErrorMessageText("Ошибка получения комнат.\n", u)
		}
		_, _ = b.view.ShowRoomsInline(rooms, u)

	case u.HasActionOrChain(view.ActionCreateRoom):
		b.HandleAddRoom(u)
	}
	return nil
}

func (b *BotApp) handleRevoteTaskRate(u *tgbot.Update) error {
	roomId := u.GetButton().GetData("roomId")
	taskId := u.GetButton().GetData("taskId")

	err := b.rateService.DelRatesByTaskId(taskId)
	if err != nil {
		return err
	}
	room, err := b.roomService.GetRoomById(roomId)
	if err != nil {
		return err
	}
	b.postTask(u, room.ChatId, taskId, roomId)
	return nil
}

func (b *BotApp) handleFinishRoom(u *tgbot.Update) error {
	roomId := u.GetButton().GetData("roomId")
	room, err := b.roomService.GetRoomById(roomId)
	if err != nil {
		log.Printf("[ERROR] unable to get room by roomId: %d, %v", roomId, err)
		return err
	}
	if room.Status == model.Finished {
		_, _ = b.view.ErrorMessage(u, "❗️ Планирование уже завершено")
		return nil
	}

	_, err = b.view.ShowTasksAfterFinishedRoom(roomId, u)
	if err == nil {
		if err = b.roomService.SetStatusRoom(model.Finished, roomId); err != nil {
			log.Printf("[ERROR] unable to set finished for room: %d, %v", roomId, err)
			return err
		}
		_, _ = b.view.ErrorMessage(u, "Планирование успешно завершено")
	} else {
		_, _ = b.view.ErrorMessage(u, "Не удалось завершить планирование")
	}
	return nil
}

func (b *BotApp) handleNextTask(u *tgbot.Update) error {
	roomId := u.GetButton().GetData("roomId")
	task, err := b.taskService.GetNextNotFinishedTask(roomId)
	if err != nil {
		log.Printf("[ERROR]  %v", err)
		_, _ = b.view.ErrorMessage(u, "❗️ Не найдено запланированных задач!")
		return nil
	}
	room, err := b.roomService.GetRoomById(roomId)
	if err != nil {
		log.Printf("[ERROR] unable to get room by roomId: %d, %v", roomId, err)
		return err
	}
	if room.UserId != u.GetUserId() {
		_, _ = b.view.ErrorMessage(u, "❗")
		return nil
	}
	b.postTask(u, room.ChatId, task.Id.String(), roomId)
	return nil
}

func (b *BotApp) handleAddRate(u *tgbot.Update) error {
	roomId := u.GetButton().GetData("roomId")
	taskId := u.GetButton().GetData("taskId")
	sum := u.GetButton().GetData("sum")
	parse, _ := uuid.Parse(taskId)

	sumInt64, err := strconv.ParseInt(sum, 10, 32)
	if err != nil {
		log.Printf("[ERROR] failed to send health, %v", err)
		return err
	}

	rate := model.Rate{
		Id:          uuid.New(),
		UserId:      u.GetUserId(),
		TaskId:      parse,
		Sum:         int32(sumInt64),
		CreatedDate: time.Now(),
	}
	if err = b.rateService.UpsertRate(rate); err != nil {
		log.Printf("[ERROR] failed to send health, %v", err)
		return err
	}

	finished, err := b.taskService.TaskFinished(taskId)
	if err != nil {
		log.Printf("[ERROR] failed to send health, %v", err)
		return err
	}

	if finished {
		rates, err := b.rateService.GetRatesByTaskId(taskId)
		if err != nil {
			log.Printf("[ERROR] unable to GetRatesByTaskId for taskId %d, %v", taskId, err)
			return err
		}
		_, _ = b.view.ShowFinishedTaskView(taskId, roomId, rates, u)
		_, _ = b.view.ShowSetTaskGrade(taskId, roomId, u)

		err = b.taskService.SetFinished(taskId)
		if err != nil {
			log.Printf("[ERROR] failed to send health, %v", err)
			return err
		}

	} else {
		//_, _ = b.view.ShowTaskTime(taskId, roomId, u)
		_, _ = b.view.ShowTaskView(0, taskId, roomId, u)
	}
	return nil
}

func (b *BotApp) postTask(u *tgbot.Update, chatId int64, taskId, roomId string) {
	msg, err := b.view.ShowTaskView(chatId, taskId, roomId, u)
	if err != nil {
		_, _ = b.view.ErrorMessage(u, "❗️ Не получилось опубликовать задачу")
	} else {
		chatIdForLink := strconv.FormatInt(chatId, 10)[4:]
		messageLink := fmt.Sprintf("[Ссылка на задачу](https://t.me/c/%v/%v) \n\n", chatIdForLink, msg.MessageID)
		go b.view.ErrorMessage(u, "Задача успешно опубликована")
		_, _ = b.view.ShowRoomView(messageLink, roomId, u)
	}
}
