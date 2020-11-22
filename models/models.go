package models

import (
	"context"

	"github.com/jackc/pgx/v4"
)

// DB holds the database connection for all models
var DB *pgx.Conn

// InitDB initilizes the connection to the database and sets DB for use.
func InitDB(databaseURL string) error {
	conn, err := pgx.Connect(context.Background(), databaseURL)
	if err != nil {
		return err
	}
	DB = conn
	return nil
}
