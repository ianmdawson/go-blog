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

// DatabaseURL retrieves DATABASE_URL from environment variables, this will be the default database url for the environment
func DatabaseURL() (string, error) {
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return "", errors.New("No DATABASE_URL provided, is it set as an environment variable?")
	}
	return databaseURL, nil
}

// InitDB initilizes the connection to the database and sets DB for use
func InitDB(databaseURL string) error {
	if databaseURL == "" {
		defaultDatabaseURL, err := DatabaseURL()
		if err != nil {
			return err
		}
		databaseURL = defaultDatabaseURL
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

// TODO: resetting the database this way doesn't work when multiple test files need to run. Update tests to mock database connection instead.

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
