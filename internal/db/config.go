package db

import (
	"context"
	"database/sql"
	"strconv"
)

func GetTelegramOffset(ctx context.Context, pgDB *sql.DB) (int64, error) {
	query := `
		SELECT value
		FROM bot_config
		WHERE key = $1;
	`

	row := pgDB.QueryRowContext(ctx, query, "telegram_offset")

	var rawValue string

	if err := row.Scan(&rawValue); err != nil {
		return 0, err
	}

	offset, convErr := strconv.ParseInt(rawValue, 10, 64)

	if convErr != nil {
		return 0, convErr
	}

	return offset, nil
}

func UpdateBotConfig(ctx context.Context, pgDB *sql.DB, key string, value int64) error {
	query := `
		UPDATE bot_config
		SET value = $1
		WHERE key = $2;
	`

	_, err := pgDB.ExecContext(ctx, query, strconv.Itoa(int(value)), key)

	if err != nil {
		return err
	}

	return nil
}
