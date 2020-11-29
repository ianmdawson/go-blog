package models

import (
	"fmt"
	"os/exec"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/assert"
)

/*
Functional integration tests require that the `blog_test` database has already been created.
If you haven't already, run the following to set up the database:
	$ make db-setup
	$ make migrate
*/

// TODO: dockerize tests and test setup
func setUpDB() {
	cmd := exec.Command("make", "-C", "../", "reset-db")
	fmt.Println("Resetting the database...")
	err := cmd.Run()
	if err != nil {
		panic(fmt.Sprint("Failed to rest the database:", err))
	}

	databaseURL := "postgres://goblog:password@localhost:5432/blog_test"
	err = InitDB(databaseURL)
	if err != nil {
		panic(fmt.Sprint("Could not connect to database", err))
	}
}

const testTitle string = "Test Page Title"
const testPageBody string = "This is a test"

func seedDatabase(t *testing.T) []*Page {
	uuid, err := uuid.NewV4()
	assert.NoError(t, err)
	page := &Page{ID: uuid, Body: []byte(testPageBody), Title: testTitle}
	err = page.Create()
	assert.NoError(t, err)
	pages := []*Page{page}
	return pages
}

func TestGetAllPages(t *testing.T) {
	setUpDB()
	createdPages := seedDatabase(t)

	offset := 0
	limit := 50
	pages, err := GetAllPages(offset, limit)
	assert.NoError(t, err)
	assert.Len(t, pages, 1)
	p := pages[0]
	assert.Equal(t, p.Title, testTitle)
	assert.Equal(t, string((p.Body)), testPageBody)
	assert.NotNil(t, p.CreatedAt)
	assert.NotNil(t, p.UpdatedAt)
	assert.Equal(t, p.ID, createdPages[0].ID)
}
