package view

import (
	"fmt"
	"github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"gotestbot/internal/service/model"
	tgbot2 "gotestbot/sdk/tgbot"
	"math"
	"sort"
	"strconv"
	"time"
)

func (v *View) AddTaskName(u *tgbot2.Update) (tgbotapi.Message, error) {
	builder := new(tgbot2.MessageBuilder).
		Message(u.GetUserId(), u.GetMessageId()).
		Text("Введите название задачи")

	return logIfError(v.tg.Send(builder.Build()))
}

func (v *View) AddTaskUrl(u *tgbot2.Update) (tgbotapi.Message, error) {
	builder := new(tgbot2.MessageBuilder).
		NewMessage(u.GetUserId()).
		Text("Введите ссылку на задачу")

	return logIfError(v.tg.Send(builder.Build()))
}

func (v *View) AddSettingTask(prefix string, u *tgbot2.Update) (tgbotapi.Message, error) {
	saveAndSendBtn := v.createButton(ActionSaveAndSendTask, nil)
	saveAndNewBtn := v.createButton(ActionSaveAndSaveTask, nil)
	saveAndCancelBtn := v.createButton(ActionSaveTaskAndCancel, nil)

	builder := new(tgbot2.MessageBuilder).
		NewMessage(u.GetUserId()).
		Text(prefix+"Выберите настройка для задачи").
		AddKeyboardRow().AddButton("💾 Сохранить и 🏹 опубликовать", saveAndSendBtn.Id).
		AddKeyboardRow().AddButton("💾 Сохранить и 🔄 создать еще", saveAndNewBtn.Id).
		AddKeyboardRow().AddButton("💾 Сохранить и выйти", saveAndCancelBtn.Id)

	return logIfError(v.tg.Send(builder.Build()))
}

func (v *View) ShowTaskView(chatId int64, taskId string, roomId string, u *tgbot2.Update) (tgbotapi.Message, error) {
	messageBuilder := new(tgbot2.MessageBuilder)
	if chatId != 0 {
		messageBuilder.NewMessage(chatId)
	} else {
		messageBuilder.Message(u.GetChatId(), u.GetMessageId()).
			Edit(u.IsButton())
	}

	room, err := v.roomProv.GetRoomById(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetRoomById for roomId: %d", roomId)
		return tgbotapi.Message{}, err
	}
	text := fmt.Sprintf("Комната: *%s*\n", room.Name)

	task, err := v.taskProv.GetTaskById(taskId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetTaskById for taskId: %d", taskId)
		return tgbotapi.Message{}, err
	}
	text += fmt.Sprintf("Задача: *%s*\n\n", task.Name)

	users, err := v.roomProv.GetUsersByRoomId(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetUsersByRoomId for roomId: %d, %v", roomId, err)
		return tgbotapi.Message{}, err
	}

	rates, err := v.rateProv.GetRatesByTaskId(taskId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetTaskById for taskId: %d, %v", taskId, err)
		return tgbotapi.Message{}, err
	}

	userIdToRate := map[int64]*model.Rate{}
	for _, rate := range rates {
		userIdToRate[rate.UserId] = &rate
	}

	for _, user := range users {
		rate := userIdToRate[user.UserId]
		rateEmoji := "🐐"
		if rate != nil && !task.Finished {
			rateEmoji = "✅"
		} else if rate != nil && task.Finished {
			rateEmoji = string(rate.Sum)
		}
		text += fmt.Sprintf("%s - %s\n", rateEmoji, userLink(&user))
	}

	messageBuilder.Text(text)
	callbackQuery := u.Update.CallbackQuery
	if callbackQuery != nil && callbackQuery.Message != nil && callbackQuery.Message.ReplyMarkup != nil &&
		callbackQuery.Message.ReplyMarkup.InlineKeyboard != nil &&
		callbackQuery.Message.ReplyMarkup.InlineKeyboard[0][0].Text == "0" {

		messageBuilder.AddKeyboard(callbackQuery.Message.ReplyMarkup.InlineKeyboard)

	} else {
		finishBtn := v.createButton(ActionFinishTask, map[string]string{"taskId": taskId, "roomId": roomId})
		messageBuilder.
			AddKeyboardRow().
			AddButton("☕️", v.createButton(ActionAddRate, map[string]string{"sum": "0", "taskId": taskId, "roomId": roomId}).Id).
			AddButton("0", v.createButton(ActionAddRate, map[string]string{"sum": "0", "taskId": taskId, "roomId": roomId}).Id).
			AddButton("1", v.createButton(ActionAddRate, map[string]string{"sum": "1", "taskId": taskId, "roomId": roomId}).Id).
			AddButton("2", v.createButton(ActionAddRate, map[string]string{"sum": "2", "taskId": taskId, "roomId": roomId}).Id).
			AddButton("3", v.createButton(ActionAddRate, map[string]string{"sum": "3", "taskId": taskId, "roomId": roomId}).Id).
			AddButton("5", v.createButton(ActionAddRate, map[string]string{"sum": "5", "taskId": taskId, "roomId": roomId}).Id).
			AddButton("8", v.createButton(ActionAddRate, map[string]string{"sum": "8", "taskId": taskId, "roomId": roomId}).Id).
			AddKeyboardRow().AddButton("Раскрыться", finishBtn.Id)
	}

	return logIfError(v.tg.Send(messageBuilder.Build()))
}

func (v *View) ShowFinishedTaskView(taskId string, roomId string, rates []model.Rate, u *tgbot2.Update) (tgbotapi.Message, error) {
	room, err := v.roomProv.GetRoomById(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetRoomById for roomId: %d, %v", roomId, err)
		return tgbotapi.Message{}, err
	}
	text := fmt.Sprintf("Комната: *%s*\n", room.Name)

	task, err := v.taskProv.GetTaskById(taskId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetTaskById for taskId: %d, %v", taskId, err)
		return tgbotapi.Message{}, err
	}
	text += fmt.Sprintf("Задача: *%s*\n\n", task.Name)

	users, err := v.roomProv.GetUsersByRoomId(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetUsersByRoomId for roomId: %d, %v", roomId, err)
		return tgbotapi.Message{}, err
	}

	var sumRates []int32
	userIdToRate := map[int64]model.Rate{}
	for _, rate := range rates {
		userIdToRate[rate.UserId] = rate
		sumRates = append(sumRates, rate.Sum)
	}

	text += "Оценки: \n"
	for _, user := range users {
		rate := userIdToRate[user.UserId]
		rateEmoji := "❓"
		if (rate != model.Rate{}) {
			rateEmoji = strconv.Itoa(int(rate.Sum))
		}
		text += fmt.Sprintf("%s - %s\n", rateEmoji, userLink(&user))
	}
	text += fmt.Sprintf("\nМедиана - *%d*", calcMedian(sumRates))

	mode, err := v.rateProv.GetModeByTaskId(taskId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetTaskById for taskId: %d, %v", taskId, err)
		return tgbotapi.Message{}, err
	}
	text += fmt.Sprintf("\nМода - *%d*", mode)

	finishBtn := v.createButton(ActionNextTask, map[string]string{"roomId": roomId})
	builder := new(tgbot2.MessageBuilder).
		Message(u.GetChatId(), u.GetMessageId()).
		Edit(u.IsButton()).
		Text(text).AddKeyboardRow().AddButton("🔜 Следующая задача", finishBtn.Id)

	return logIfError(v.tg.Send(builder.Build()))
}

func calcMedian(sums []int32) int32 {
	if sums == nil || len(sums) == 0 {
		return 0
	}
	sort.Slice(sums, func(i, j int) bool { return sums[i] < sums[j] })
	mNumber := len(sums) / 2

	if isOdd(sums) {
		return sums[mNumber]
	}
	return (sums[mNumber-1] + sums[mNumber]) / 2
}

func isOdd(sums []int32) bool {
	if len(sums)%2 == 0 {
		return false
	}
	return true
}

func (v *View) ShowTaskTime(taskId string, roomId string, u *tgbot2.Update) (tgbotapi.Message, error) {
	task, err := v.taskProv.GetTaskById(taskId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetTaskById for taskId: %d, %v", taskId, err)
		return tgbotapi.Message{}, err
	}
	room, err := v.roomProv.GetRoomById(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetRoomById for roomId: %d, %v", roomId, err)
		return tgbotapi.Message{}, err
	}
	if !task.Finished {
		print(room.ChatId)
		sub := task.CreatedDate.Sub(time.Now())
		return v.WarnMessage(fmt.Sprintf("⏳ Осталось %v минут, %v секунд", sub.Minutes(), sub.Seconds()), u)
	}
	return tgbotapi.Message{}, nil
}

func (v *View) ShowTasks(roomId string, page int, u *tgbot2.Update) (tgbotapi.Message, error) {
	room, err := v.roomProv.GetRoomById(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetRoomById for roomId: %d, %v", roomId, err)
		return tgbotapi.Message{}, err
	}
	tasks, err := v.taskProv.GetTasksByRoomIdAndPagination(room.Id.String(), page*10, 10)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetTasksByRoomId for roomId: %d, %v", roomId, err)
		return tgbotapi.Message{}, err
	}
	if tasks == nil {
		_, _ = v.ErrorMessage(u, "❗️ Не найдены задачи")
		return tgbotapi.Message{}, nil
	}

	builder := new(tgbot2.MessageBuilder).
		Message(u.GetChatId(), u.GetMessageId()).
		Edit(u.IsButton()).Text(fmt.Sprintf("Задачи в комнате: *%v*", room.Name))
	for _, task := range tasks {
		taskBtn := v.createButton(ActionShowTask, map[string]string{"taskId": task.Id.String(), "roomId": task.RoomId.String()})
		finishedEmoji := "❌"
		if task.Finished {
			finishedEmoji = "✅ " + strconv.FormatInt(int64(task.Grade), 32)
		}
		builder.AddKeyboardRow().AddButton(fmt.Sprintf("%v %v", finishedEmoji, task.Name), taskBtn.Id)
	}

	backBtn := v.createButton(ActionShowRoom, map[string]string{"roomId": roomId})
	builder.AddKeyboardRow()
	shwTasksBtn := v.createButton(ActionShowTasks, map[string]string{"roomId": roomId, "page": strconv.Itoa(page - 1)})
	shwTasksNext := v.createButton(ActionShowTasks, map[string]string{"roomId": roomId, "page": strconv.Itoa(page + 1)})

	builder.AddButton("⬅️", shwTasksBtn.Id).
		AddButton("Назад", backBtn.Id).
		AddButton("➡️️", shwTasksNext.Id)

	return logIfError(v.tg.Send(builder.Build()))
}

func (v *View) ShowTasksAfterFinishedRoom(roomId string, u *tgbot2.Update) (tgbotapi.Message, error) {
	room, err := v.roomProv.GetRoomById(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetRoomById for roomId: %d, %v", roomId, err)
		return tgbotapi.Message{}, err
	}
	tasks, err := v.taskProv.GetTasksByRoomIdAndPagination(room.Id.String(), 0, math.MaxInt64)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetTasksByRoomId for roomId: %d, %v", roomId, err)
		return tgbotapi.Message{}, err
	}
	if tasks == nil {
		_, _ = v.ErrorMessage(u, "❗️ Не найдены задачи")
		return tgbotapi.Message{}, nil
	}

	text := fmt.Sprintf("*❗ Планирование завершено*\n\nКомната: *%v*\nЗадачи:\n", room.Name)
	for _, task := range tasks {
		text += fmt.Sprintf("- *%v* %v\n", task.Grade, task.Name)
	}

	builder := new(tgbot2.MessageBuilder).
		NewMessage(room.ChatId).
		Text(text)
	return logIfError(v.tg.Send(builder.Build()))
}

func (v *View) ShowSetTaskGrade(taskId, roomId string, u *tgbot2.Update) (tgbotapi.Message, error) {
	room, err := v.roomProv.GetRoomById(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetRoomById for roomId: %d, $v", roomId, err)
		return tgbotapi.Message{}, err
	}
	text := fmt.Sprintf("Комната: *%s*\n\n", room.Name)

	task, err := v.taskProv.GetTaskById(taskId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetTaskById for taskId: %d, $v", taskId, err)
		return tgbotapi.Message{}, err
	}
	text += fmt.Sprintf("Завершена оценка по задаче: *%s*\n", task.Name)

	sumRates, err := v.rateProv.GetRatesSums(taskId)
	if err != nil {
		lgr.Printf("[ERROR] unable to GetRatesByTaskId for taskId: %d, $v", taskId, err)
		return tgbotapi.Message{}, err
	}

	if sumRates != nil {
		text += fmt.Sprintf("\nМедиана - *%d*", calcMedian(sumRates))

		mode, err := v.rateProv.GetModeByTaskId(taskId)
		if err != nil {
			lgr.Fatalf("[ERROR] unable to GetRatesByTaskId for taskId: %d, $v", taskId, err)
			return tgbotapi.Message{}, err
		}
		text += fmt.Sprintf("\nМода - *%d*", mode)
	}

	finishRateBtn := v.createButton(ActionFinishTaskRate, map[string]string{"roomId": roomId, "taskId": taskId})
	revoteRateBtn := v.createButton(ActionRevoteTaskRate, map[string]string{"roomId": roomId, "taskId": taskId})

	builder := new(tgbot2.MessageBuilder).
		NewMessage(room.UserId).
		Text(text).
		AddKeyboardRow().AddButton("Ввести итоговую оценку", finishRateBtn.Id).
		AddKeyboardRow().AddButton("Переголосовать", revoteRateBtn.Id)

	return logIfError(v.tg.Send(builder.Build()))

}
