package db

import (
	"database/sql"
	"log"
)

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
