package model

import (
	"github.com/google/uuid"
	"time"
)

type RoomStatus string

const (
	New      = RoomStatus("NEW")
	Finished = RoomStatus("FINISHED")
)

type Room struct {
	Id          uuid.UUID  `db:"id"`
	Status      RoomStatus `db:"status"`
	Name        string     `db:"name"`
	UserId      int64      `db:"user_id"`
	ChatId      int64      `db:"chat_id"`
	CreatedDate time.Time  `db:"created_date"`
}

type Task struct {
	Id          uuid.UUID `db:"id"`
	Name        string    `db:"name"`
	Url         string    `db:"url"`
	RoomId      uuid.UUID `db:"room_id"`
	Grade       int32     `db:"grade"`
	Finished    bool      `db:"finished"`
	CreatedDate time.Time `db:"created_date"`
}

type Rate struct {
	Id          uuid.UUID `db:"id"`
	UserId      int64     `db:"user_id"`
	TaskId      uuid.UUID `db:"task_id"`
	Sum         int32     `db:"sum"`
	CreatedDate time.Time `db:"created_date"`
}
