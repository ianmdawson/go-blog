package models

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

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

// TTearDown test helper method that handles closes the database connection
// after running SetUpDB, simply add the following line:
// defer TearDown()
func TTearDown() {
	DB.Close(context.Background())
	return
}

// TSetUpDB test helper method that resets the test database, handles connecting to the test database
func TSetUpDB() {
	cmd := exec.Command("make", "-C", "../", "reset-db-test")
	fmt.Println("Resetting the test database...")
	err := cmd.Run()
	if err != nil {
		panic(fmt.Sprint("Failed to reset the database:", err))
	}

	databaseURL := "postgres://goblog:password@localhost:5432/blog_test"
	err = InitDB(databaseURL)
	if err != nil {
		panic(fmt.Sprint("Could not connect to database", err))
	}
}
