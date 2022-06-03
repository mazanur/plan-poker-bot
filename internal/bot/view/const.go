package view

import (
	"gotestbot/sdk/tgbot"
)

const (
	ActionStart = tgbot.Action("START")
)

const (
	ActionCancel            = tgbot.Action("CANCEL")
	ActionCreateRoom        = tgbot.Action("NEW_ROOM")
	ActionShowRooms         = tgbot.Action("SHOW_ROOMS")
	ActionShowRoom          = tgbot.Action("SHOW_ROOM")
	ActionJoinRoom          = tgbot.Action("JOIN_ROOM")
	ActionBotAdded          = tgbot.Action("BOT_ADDED")
	ActionSetGroupOfRoom    = tgbot.Action("SET_GROUP_OF_ROOM")
	ActionFinishRoom        = tgbot.Action("FINISH_ROOM")
	ActionRoomSettingTimes  = tgbot.Action("SETTINGS_ROOM_TIMER")
	ActionCreateTask        = tgbot.Action("ADD_TASK")
	ActionShowTasks         = tgbot.Action("SHOW_TASKS")
	ActionShowTask          = tgbot.Action("SHOW_TASK")
	ActionNextTask          = tgbot.Action("NEXT_TASK")
	ActionSaveAndSendTask   = tgbot.Action("SAVE_AND_SEND_TASK")
	ActionSaveAndSaveTask   = tgbot.Action("SAVE_AND_NEW_TASK")
	ActionSaveTaskAndCancel = tgbot.Action("SAVE_TASK_AND_CANCEL")
	ActionFinishTask        = tgbot.Action("FINISH_TASK")
	ActionAddRate           = tgbot.Action("TASK_RATE")
	ActionRevoteTaskRate    = tgbot.Action("REVOTE_TASK_RATE")
	ActionFinishTaskRate    = tgbot.Action("FINISH_TASK_RATE")
)
