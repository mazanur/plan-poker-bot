package bot_handler

import (
	"github.com/rs/zerolog/log"
	"gotestbot/internal/bot/view"
	"gotestbot/sdk/tgbot"
	"strconv"
)

func (b *BotApp) HandleAddTaskGrade(u *tgbot.Update) {

	if u.HasAction(view.ActionFinishTaskRate) {
		roomId := u.GetButton().GetData("roomId")
		taskId := u.GetButton().GetData("taskId")

		u.StartChain(string(view.ActionFinishTaskRate)).
			StartChainStep("SET_GRADE").
			AddChainData("roomId", roomId).
			AddChainData("taskId", taskId).
			FlushChatInfo()

		_, _ = b.view.ErrorMessageText("Введите итоговую оценку", u)
		return
	}

	switch u.GetChainStep() {
	case "SET_GRADE":
		grade := u.GetText()
		gradeInt64, err := strconv.ParseInt(grade, 10, 32)
		if err != nil {
			log.Error().Err(err).Msgf("unable to get room by roomId: %d", grade)
		}

		taskId := u.GetChainData("taskId")
		if err = b.taskService.SetGradeTask(int32(gradeInt64), taskId); err != nil {
			log.Error().Err(err).Msgf("unable SetGradeTask by taskId: %v", taskId)
			_, _ = b.view.ErrorMessageText("❗️ Ошибка присваивания итоговой оценки задаче", u)
			return
		}
		roomId := u.GetChainData("roomId")
		_, _ = b.view.ShowRoomView("Итоговая оценка успешно присвоена\n\n", roomId, u)
		u.FinishChain().FlushChatInfo()

	default:
		_, _ = b.view.ErrorMessageText("❗️ Ошибка присваивания итоговой оценки задаче", u)
	}

}
