package models

import (
	"context"
	"errors"
	"os"

	"github.com/jackc/pgx/v4"
)

// DB holds the database connection for all models
var DB *pgx.Conn

// InitDB initilizes the connection to the database and sets DB for use
func InitDB(databaseURL string) error {
	if databaseURL == "" {
		databaseURL = os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			return errors.New("No DATABASE_URL provided, is it set as an environment variable?")
		}
	}

	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return err
	}
	DB = conn
	return nil
}
