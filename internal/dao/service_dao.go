package dao

import (
	"database/sql"
	"github.com/pkg/errors"
	"gotestbot/internal/service/model"
	"gotestbot/sdk/tgbot"
)

func (r *Repository) SaveTask(task model.Task) error {
	insert := `INSERT INTO task(id, name, url, room_id, finished, created_date, grade) VALUES (:id, :name, :url, :room_id, :finished, :created_date, :grade)`

	if _, err := r.db.NamedExec(insert, task); err != nil {
		return err
	}
	return nil
}

func (r *Repository) SetFinishedTask(taskId string) error {
	_, err := r.db.Exec(`UPDATE task SET finished = TRUE WHERE id = $1;`, taskId)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) SetGradeTask(grade int32, taskId string) error {
	_, err := r.db.Exec(`UPDATE task SET grade = $1 WHERE id = $2;`, grade, taskId)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetTaskById(taskId string) (model.Task, error) {
	row := r.db.QueryRowx("SELECT * FROM task WHERE id = $1", taskId)

	task := model.Task{}
	if err := row.StructScan(&task); err != nil {
		return model.Task{}, errors.Wrapf(err, "unable to get room, roomId: %v", taskId)
	}
	return task, nil
}

func (r *Repository) GetTasksByRoomId(roomId string, offset, limit int) ([]model.Task, error) {
	rows, err := r.db.Queryx(`SELECT *FROM task WHERE room_id = $1 ORDER BY created_date  LIMIT $2 OFFSET $3`, roomId, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var tasks []model.Task
	for rows.Next() {
		t := model.Task{}
		if err = rows.StructScan(&t); err != nil {
			return []model.Task{}, errors.Wrapf(err, "unable to get tasks, roomId: %v", roomId)
		}
		tasks = append(tasks, t)
	}
	return tasks, nil
}

func (r *Repository) GetNextNotFinishedTask(roomId string) (model.Task, error) {
	const query = `SELECT * FROM task 
				   WHERE finished IS FALSE AND room_id = $1 
			       ORDER BY created_date LIMIT 1`
	row := r.db.QueryRowx(query, roomId)

	task := model.Task{}
	if err := row.StructScan(&task); err != nil {
		return model.Task{}, errors.Wrapf(err, "unable to get next not finished task, roomId: %v", roomId)
	}
	return task, nil
}

func (r *Repository) TaskFinished(taskId string) (bool, error) {
	var finished bool
	row := r.db.QueryRow(`SELECT(SELECT count(1) = (SELECT count(1)
									  FROM room_member rm
									  WHERE rm.room_id = (SELECT t.room_id FROM task t WHERE t.id = $1 )) 
							FROM rate r
							WHERE task_id = $1)`, taskId)
	err := row.Scan(&finished)
	if err != nil {
		return false, err
	}
	return finished, nil
}

func (r *Repository) SaveRoom(room model.Room) error {
	insert := `INSERT INTO room(id, name, user_id, status, chat_id, created_date) VALUES (:id, :name, :user_id, :status, :chat_id, :created_date)`

	if _, err := r.db.NamedExec(insert, room); err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetRoomById(roomId string) (model.Room, error) {
	row := r.db.QueryRowx("SELECT * FROM room WHERE id = $1", roomId)

	room := model.Room{}
	if err := row.StructScan(&room); err != nil {
		return model.Room{}, errors.Wrapf(err, "unable to get room, roomId: %v", roomId)
	}
	return room, nil
}

func (r *Repository) SetChatIdRoom(roomId string, chatId int64) error {
	_, err := r.db.Exec(`UPDATE room SET chat_id = $2 WHERE id = $1;`, roomId, chatId)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) SetStatusRoom(status model.RoomStatus, roomId string) error {
	_, err := r.db.Exec(`UPDATE room SET status = $1 WHERE id = $2;`, status, roomId)
	if err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetRoomsByNameAndUserId(roomName string, userId int64) ([]model.Room, error) {
	query := `SELECT * FROM room r
 			  WHERE name ILIKE $1 AND user_id = $2`
	rows, err := r.db.Queryx(query, roomName, userId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rooms []model.Room
	for rows.Next() {
		r := model.Room{}
		if err = rows.StructScan(&r); err != nil {
			return []model.Room{}, errors.Wrapf(err, "unable to get tasks, userId :%v, and roomName: %v", userId, roomName)
		}
		rooms = append(rooms, r)
	}
	return rooms, nil
}

func (r *Repository) GetUsersByRoomId(roomId string) ([]tgbot.User, error) {
	rows, err := r.db.Queryx(`SELECT p.* FROM profile p 
    							JOIN room_member rm ON rm.user_id = p.user_id 
								WHERE rm.room_id = $1`, roomId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []tgbot.User
	for rows.Next() {
		u := tgbot.User{}
		if err = rows.StructScan(&u); err != nil {
			return []tgbot.User{}, errors.Wrapf(err, "unable to get users, roomId: %d", roomId)
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *Repository) GetRoom(roomId string) ([]tgbot.User, error) {
	rows, err := r.db.Queryx(`SELECT p.* FROM profile p 
    							JOIN room_member rm ON rm.user_id = p.user_id 
								WHERE rm.room_id = $1`, roomId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var users []tgbot.User
	for rows.Next() {
		u := tgbot.User{}
		if err = rows.StructScan(&u); err != nil {
			return []tgbot.User{}, errors.Wrapf(err, "unable to get users, roomId: %d", roomId)
		}
		users = append(users, u)
	}

	return users, nil
}

func (r *Repository) SaveRoomMember(userId int64, roomId string) error {
	insert := `INSERT INTO room_member(user_id, room_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`

	if _, err := r.db.Exec(insert, userId, roomId); err != nil {
		return err
	}
	return nil
}

func (r *Repository) SaveRate(rate model.Rate) error {
	insert := `INSERT INTO rate(id, task_id, user_id, sum, created_date) VALUES (:id, :task_id, :user_id, :sum, :created_date)`

	if _, err := r.db.NamedExec(insert, rate); err != nil {
		return err
	}
	return nil
}

func (r *Repository) UpdateRate(rateId string, rate model.Rate) error {
	insert := `UPDATE rate SET sum = $2 WHERE id = $1`
	if _, err := r.db.Query(insert, rateId, rate.Sum); err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetRateByUserAndTaskId(userId int64, taskId string) (*model.Rate, error) {
	row := r.db.QueryRowx("SELECT rate.* FROM rate WHERE task_id = $1 AND user_id = $2", taskId, userId)
	rate := model.Rate{}
	err := row.StructScan(&rate)
	if err == sql.ErrNoRows {
		return nil, nil
	} else if err != nil {
		return &rate, err
	}
	return &rate, nil
}

func (r *Repository) GetRatesByTaskId(taskId string) ([]model.Rate, error) {
	rows, err := r.db.Queryx(`SELECT r.* FROM rate r 
								WHERE r.task_id = $1`, taskId)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var rates []model.Rate
	for rows.Next() {
		r := model.Rate{}
		if err = rows.StructScan(&r); err != nil {
			return []model.Rate{}, errors.Wrapf(err, "unable to get rates, taskId: %v", taskId)
		}
		rates = append(rates, r)
	}

	return rates, nil
}

func (r *Repository) DelRatesByTaskId(taskId string) error {
	query := "DELETE FROM rate WHERE task_id = $1"
	if _, err := r.db.Query(query, taskId); err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetModeByTaskId(taskId string) (int32, error) {
	var mode int32
	row := r.db.QueryRow("SELECT mode() within GROUP (order by sum) FROM rate WHERE task_id = $1", taskId)
	err := row.Scan(&mode)
	if err != nil {
		return 0, err
	}
	return mode, nil
}
