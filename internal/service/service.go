package service

import (
	"github.com/pkg/errors"
	"gotestbot/internal/dao"
	"gotestbot/internal/service/model"
)

type RoomService struct {
	*dao.Repository
}

func NewRoomService(repository *dao.Repository) *RoomService {
	return &RoomService{Repository: repository}
}

type TaskService struct {
	r *dao.Repository
}

func NewTaskService(repository *dao.Repository) *TaskService {
	return &TaskService{r: repository}
}

func (s TaskService) SaveTask(task model.Task) error {
	return s.r.SaveTask(task)
}

func (s TaskService) SetFinished(taskId string) error {
	return s.r.SetFinishedTask(taskId)
}

func (s TaskService) GetTaskById(taskId string) (model.Task, error) {
	return s.r.GetTaskById(taskId)
}

func (s TaskService) GetTasksByRoomId(roomId string) ([]model.Task, error) {
	return s.r.GetTasksByRoomId(roomId, 0, 100)
}

func (s TaskService) GetTasksByRoomIdAndPagination(roomId string, offset, limit int) ([]model.Task, error) {
	return s.r.GetTasksByRoomId(roomId, offset, limit)
}

func (s TaskService) GetNextNotFinishedTask(roomId string) (model.Task, error) {
	return s.r.GetNextNotFinishedTask(roomId)
}

func (s TaskService) TaskFinished(taskId string) (bool, error) {
	return s.r.TaskFinished(taskId)
}

func (s TaskService) SetGradeTask(grade int32, taskId string) error {
	return s.r.SetGradeTask(grade, taskId)
}

type RateService struct {
	r *dao.Repository
}

func NewRateService(repository *dao.Repository) *RateService {
	return &RateService{r: repository}
}

func (s RateService) GetRatesByTaskId(taskId string) ([]model.Rate, error) {
	return s.r.GetRatesByTaskId(taskId)
}

func (s RateService) DelRatesByTaskId(taskId string) error {
	return s.r.DelRatesByTaskId(taskId)
}

func (s RateService) GetRatesSums(taskId string) ([]int32, error) {
	rates, err := s.r.GetRatesByTaskId(taskId)
	if err != nil {
		return []int32{}, err
	}
	var sumRates []int32
	for _, rate := range rates {
		sumRates = append(sumRates, rate.Sum)
	}
	return sumRates, nil

}

func (s RateService) UpsertRate(rate model.Rate) error {
	rateByUser, err := s.r.GetRateByUserAndTaskId(rate.UserId, rate.TaskId.String())
	if err != nil {
		return errors.Wrapf(err, "cannot querying for get rates")
	}
	if rateByUser == nil {
		err = s.r.SaveRate(rate)
		if err != nil {
			return errors.Wrapf(err, "cannot save rate %v", rate)
		}
	} else {
		if err = s.r.UpdateRate(rateByUser.Id.String(), rate); err != nil {
			return errors.Wrapf(err, "cannot querying for update rate. rateId=%v", rate.Id.String())
		}
	}
	return nil
}

func (s RateService) GetModeByTaskId(taskId string) (int32, error) {
	return s.r.GetModeByTaskId(taskId)
}
