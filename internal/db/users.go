package db

import (
	"context"
	"database/sql"
	"log"
)

type User struct {
	ChatId     int64  `json:"chat_id"`
	FirstName  string `json:"first_name"`
	UserName   string `json:"username,omitempty"`
	Subscribed bool   `json:"subscribed"`
}

func GetSubscribedUsers(pgDB *sql.DB) ([]int64, error) {
	query := `
		SELECT chat_id 
		FROM users
		WHERE subscribed = true;
	`

	rows, err := pgDB.Query(query)

	if err != nil {
		log.Println(err)
		return nil, err
	}

	defer rows.Close()

	var chatIDs []int64

	for rows.Next() {
		var chatId int64
		scanErr := rows.Scan(&chatId)

		if scanErr != nil {
			log.Println(scanErr)
			continue
		}

		chatIDs = append(chatIDs, chatId)
	}

	return chatIDs, nil
}

func AddNewUser(ctx context.Context, pgDB *sql.DB, user User) error {
	query := `
		INSERT INTO users (chat_id, first_name, username, subscribed)
		VALUES ($1, $2, $3, false)
		ON CONFLICT (chat_id) DO UPDATE
		SET first_name = $2,
			username = $3
	`

	_, err := pgDB.ExecContext(ctx, query, user.ChatId, user.FirstName, user.UserName)

	if err != nil {
		log.Println("Error adding new user:", err)
		return err
	}

	return nil
}

func UpdateSubscription(ctx context.Context, pgDB *sql.DB, chatID int64, subscribed bool) error {
	query := `
		UPDATE users SET subscribed = $1 WHERE chat_id = $2
	`

	_, err := pgDB.ExecContext(ctx, query, subscribed, chatID)

	if err != nil {
		log.Println("Error updating user subscription:", err)
		return err
	}

	if subscribed {
		log.Println("User subscribed:", chatID)
	} else {
		log.Println("User un-subscribed:", chatID)
	}

	return nil
}
