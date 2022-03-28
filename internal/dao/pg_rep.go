package dao

import (
	_ "github.com/jackc/pgx/v4/stdlib"
	"github.com/jmoiron/sqlx"
	"github.com/pkg/errors"
	"gotestbot/sdk/tgbot"
)

type Repository struct {
	db *sqlx.DB
}

func NewRepository(db *sqlx.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetChat(chatId int64) (chat tgbot.ChatInfo, err error) {
	row := r.db.QueryRowx("SELECT * FROM chat_info WHERE chat_id = $1", chatId)

	if err = row.StructScan(&chat); err != nil {
		return tgbot.ChatInfo{}, errors.Wrapf(err, "unable to get chatInfo, chatId: %d", chatId)
	}
	return
}

func (r *Repository) SaveChatInfo(chat tgbot.ChatInfo) error {

	insert := `INSERT INTO chat_info (chat_id, active_chain, active_chain_step, chain_data)
								VALUES (:chat_id, :active_chain, :active_chain_step, :chain_data)
								ON CONFLICT (chat_id) DO UPDATE SET active_chain      = :active_chain,
														 			active_chain_step = :active_chain_step,
														  			chain_data        = :chain_data`

	if _, err := r.db.NamedExec(insert, chat); err != nil {
		return errors.Wrap(err, "unable to save chatInfo")
	}
	return nil
}

func (r *Repository) GetButton(btnId string) (btn tgbot.Button, err error) {
	row := r.db.QueryRowx("SELECT * FROM button WHERE id = $1", btnId)

	if err = row.StructScan(&btn); err != nil {
		return tgbot.Button{}, errors.Wrapf(err, "unable to get button, btnId: %s", btnId)
	}
	return
}

func (r *Repository) SaveButton(button tgbot.Button) error {
	insert := "INSERT INTO button (id, action, data) VALUES (:id, :action, :data)"

	if _, err := r.db.NamedExec(insert, button); err != nil {
		return err
	}
	return nil
}

func (r *Repository) SaveUser(user tgbot.User) error {
	insert := `INSERT INTO profile (user_id, user_name, display_name) VALUES (:user_id, :user_name, :display_name)
					ON CONFLICT (user_id) DO UPDATE SET user_name    = :user_name,
											       		display_name = :display_name`

	if _, err := r.db.NamedExec(insert, user); err != nil {
		return err
	}
	return nil
}

func (r *Repository) GetUser(userId int64) (tgbot.User, error) {
	row := r.db.QueryRowx("SELECT * FROM profile WHERE user_id = $1", userId)

	var user = tgbot.User{}
	if err := row.StructScan(&user); err != nil {
		return tgbot.User{}, errors.Wrapf(err, "unable to get GetUser, userId: %v", user)
	}
	return user, nil
}
