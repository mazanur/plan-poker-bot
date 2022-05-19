package view

import (
	"fmt"
	"github.com/go-pkgz/lgr"
	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
	"github.com/google/uuid"
	"gotestbot/internal/service/model"
	"gotestbot/sdk/tgbot"
	"strconv"
)

type RoomProvider interface {
	GetRoomById(roomId string) (model.Room, error)
	GetUsersByRoomId(roomId string) ([]tgbot.User, error)
}

type TaskProvider interface {
	GetTaskById(taskId string) (model.Task, error)
	GetTasksByRoomId(roomId string) ([]model.Task, error)
	GetTasksByRoomIdAndPagination(roomId string, offset, limit int) ([]model.Task, error)
}

type RateProvider interface {
	GetRatesByTaskId(taskId string) ([]model.Rate, error)
	GetRatesSums(taskId string) ([]int32, error)
	UpsertRate(rate model.Rate) error
	GetModeByTaskId(taskId string) (int32, error)
}

type UserProvider interface {
	GetUser(userId int64) (tgbot.User, error)
}

type View struct {
	chatProv tgbot.ChatProvider
	userProv UserProvider
	roomProv RoomProvider
	taskProv TaskProvider
	rateProv RateProvider

	tg *tgbot.Bot
}

func NewView(btnProv tgbot.ChatProvider, userProv UserProvider, roomProv RoomProvider, taskProv TaskProvider, rateProv RateProvider, tg *tgbot.Bot) *View {
	return &View{
		chatProv: btnProv,
		userProv: userProv,
		roomProv: roomProv,
		taskProv: taskProv,
		rateProv: rateProv,
		tg:       tg}
}

func (v *View) StartView(u *tgbot.Update) (tgbotapi.Message, error) {

	crtBtn := v.createButton(ActionCreateRoom, nil)
	showBtn := v.createButton(ActionShowRooms, nil)

	msg := new(tgbot.MessageBuilder).
		Message(u.GetChatId(), u.GetMessageId()).
		Edit(u.IsButton()).
		Text("Добро пожаловать! \nЭто *PlanPokerBot*. Выберите одно из предоложенных действий").
		AddKeyboardRow().AddButton("Создать комнату", crtBtn.Id).
		AddKeyboardRow().AddButton("Просмотреть комнаты", showBtn.Id).
		Build()

	return logIfError(v.tg.Send(msg))
}

func (v *View) ShowRoomView(prefix, roomId string, u *tgbot.Update) (tgbotapi.Message, error) {
	users, err := v.roomProv.GetUsersByRoomId(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to get users by roomId: %d", roomId)
	}

	var members string
	for _, user := range users {
		members += "- " + userLink(&user) + "\n"
	}
	room, err := v.roomProv.GetRoomById(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to get room by roomId: %d", roomId)
	}

	builder := new(tgbot.MessageBuilder).
		Message(u.GetUserId(), u.GetMessageId()).
		Edit(u.IsButton()).
		Text(prefix + fmt.Sprintf("Комната - *%v*\n🗓 %v \n\nУчастники:\n%v", room.Name, room.CreatedDate.Format("02 January 2006"), members))

	backBtn := v.createButton(ActionStart, nil)
	addTaskBtn := v.createButton(ActionCreateTask, map[string]string{"roomId": roomId})
	tasksBtn := v.createButton(ActionShowTasks, map[string]string{"roomId": roomId, "page": "0"})
	nextTaskBtn := v.createButton(ActionNextTask, map[string]string{"roomId": roomId})
	finishRmBtn := v.createButton(ActionFinishRoom, map[string]string{"roomId": roomId})

	builder.AddKeyboardRow().AddButton("➕ Добавить задачу", addTaskBtn.Id).
		AddKeyboardRow().AddButtonSwitch("📢 Отправить в чат", room.Name).
		AddKeyboardRow().AddButton("🗂 Задачи", tasksBtn.Id).AddButton("📤 Следующая задача", nextTaskBtn.Id).
		AddKeyboardRow().AddButton("🏁 Завершить планирование", finishRmBtn.Id).
		AddKeyboardRow().AddButton("Назад", backBtn.Id)
	return logIfError(v.tg.Send(builder.Build()))
}

func (v *View) ErrorMessage(u *tgbot.Update, text string) (tgbotapi.Message, error) {
	c := &tgbotapi.CallbackConfig{
		CallbackQueryID: u.CallbackQuery.ID,
		Text:            text,
		ShowAlert:       true,
	}
	return logIfError(v.tg.Send(c))
}

func (v *View) WarnMessage(text string, u *tgbot.Update) (tgbotapi.Message, error) {
	c := &tgbotapi.CallbackConfig{
		CallbackQueryID: u.CallbackQuery.ID,
		Text:            text,
		ShowAlert:       false,
	}
	return logIfError(v.tg.Send(c))
}

func (v *View) ErrorMessageText(text string, u *tgbot.Update) (tgbotapi.Message, error) {
	msg := new(tgbot.MessageBuilder).
		Message(u.GetUserId(), u.GetMessageId()).
		Edit(u.IsButton()).
		Text(text).
		Build()

	return logIfError(v.tg.Send(msg))
}

func (v *View) NewDeleteMessage(chatID int64, messageID int) (tgbotapi.Message, error) {
	c := tgbotapi.NewDeleteMessage(chatID, messageID)
	return logIfError(v.tg.Send(c))
}

func (v *View) SendChatWritingAction(chatId int64) (tgbotapi.Message, error) {
	msg := tgbotapi.NewChatAction(chatId, tgbotapi.ChatTyping)
	return logIfError(v.tg.Send(msg))
}

func (v *View) ShowRoomsInline(rooms []model.Room, u *tgbot.Update) (tgbotapi.Message, error) {
	inlineRequest := tgbot.NewInlineRequest(u.GetInlineId())
	for _, room := range rooms {
		joinBtn := v.createButton(ActionJoinRoom, map[string]string{"roomId": room.Id.String()})

		users, err := v.roomProv.GetUsersByRoomId(room.Id.String())
		if err != nil {
			lgr.Printf("[ERROR] unable to get users by roomId: %d", room.Id.String())
		}

		var members string
		for _, user := range users {
			members += "- " + userLink(&user) + "\n"
		}

		inlineRequest.AddArticle(uuid.NewString(),
			room.Name, "Статус",
			fmt.Sprintf("Комната - *%v*\n🗓 %v \n\nУчастники:\n%v", room.Name, room.CreatedDate.Format("02 January 2006"), members)).
			AddKeyboardRow().AddButton("Присоединиться", joinBtn.Id)
	}

	return logIfError(v.tg.Send(inlineRequest.Build()))
}

func (v *View) ShowRoomViewInline(roomId string, u *tgbot.Update) (tgbotapi.Message, error) {
	users, err := v.roomProv.GetUsersByRoomId(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to get users by roomId: %d", roomId)
	}

	var members string
	for _, user := range users {
		members += "- " + userLink(&user) + "\n"
	}
	room, err := v.roomProv.GetRoomById(roomId)
	if err != nil {
		lgr.Printf("[ERROR] unable to get room by roomId: %d", roomId)
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

func (v *View) ChangeChatOfRoom(room model.Room, chat *tgbotapi.Chat, u *tgbot.Update) (tgbotapi.Message, error) {
	cancelBtn := v.createButton(ActionCancel, nil)
	setGroupBtn := v.createButton(ActionSetGroupOfRoom, map[string]string{
		"roomId":   room.Id.String(),
		"chatId":   strconv.FormatInt(chat.ID, 10),
		"chatName": chat.Title})

	text := fmt.Sprintf("Вы отправили сообщение в чат *%v* для привязки к комнате *%v*", chat.Title, room.Name)
	builder := new(tgbot.MessageBuilder).
		NewMessage(u.GetUserId()).
		Text(text).
		AddKeyboardRow().AddButton("🔗 Привязать", setGroupBtn.Id).
		AddKeyboardRow().AddButton("Отмена", cancelBtn.Id)

	return logIfError(v.tg.Send(builder.Build()))
}

func (v *View) GetMe() tgbotapi.User {
	me, _ := v.tg.GetMe()
	return me
}
