package models

import (
	"context"
	"time"

	"github.com/gofrs/uuid"
)

// Page represents page data
type Page struct {
	ID        uuid.UUID
	Title     string
	Body      []byte
	CreatedAt time.Time
	UpdatedAt time.Time
}

// PageCollection is a collection of Page results. Contains metadata useful for pagination
type PageCollection struct {
	Pages             []*Page
	Count             int
	ResultsPageNumber int
	Limit             int
	NextPage          int
	PreviousPage      int
	AtLastPage        bool
}

// Update the existing page in the database
func (p *Page) Update() error {
	sql := ` -- name: PageUpdate :one
		UPDATE pages
		SET title = $2, body = $3, updated_at = now()
		WHERE id=$1
		RETURNING id, title, body, created_at, updated_at
		;`
	err := DB.QueryRow(context.Background(), sql, p.ID, p.Title, p.Body).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// Create a page in the database
func (p *Page) Create() error {
	sql := ` -- name: PageCreate :one
		INSERT INTO pages
		(id, title, body)
		VALUES ($1, $2, $3)
		ON CONFLICT (id) DO NOTHING
		RETURNING id, title, body, created_at, updated_at
		;`
	err := DB.QueryRow(context.Background(), sql, p.ID, p.Title, p.Body).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// Find a page, given a UUID
func (p *Page) Find(id uuid.UUID) error {
	sql := `SELECT id, title, body, created_at, updated_at FROM pages WHERE id=$1;`
	err := DB.QueryRow(context.Background(), sql, id).Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
	if err != nil {
		return err
	}

	return nil
}

// GetAllPages retrieves all Page models in the database
// limit the number of results to return
// offset the number of results to skip
func GetAllPages(offset int, limit int) ([]*Page, error) {
	sql := `SELECT id, title, body, created_at, updated_at FROM pages ORDER BY created_at DESC OFFSET $1 LIMIT $2;`
	rows, err := DB.Query(context.Background(), sql, offset, limit)
	defer rows.Close()
	if err != nil {
		return nil, err
	}

	var pages []*Page

	for rows.Next() {
		p := &Page{}
		err = rows.Scan(&p.ID, &p.Title, &p.Body, &p.CreatedAt, &p.UpdatedAt)
		if err != nil {
			return nil, err
		}
		pages = append(pages, p)
	}

	return pages, nil
}

func GetPageCollection(offset int, limit int) (*PageCollection, error) {
	pages, err := GetAllPages(offset, limit)
	if err != nil {
		return nil, err
	}

	count, err := CountAllPages()
	if err != nil {
		return nil, err
	}

	resultsPageNumber := 1
	if offset != 0 {
		resultsPageNumber = (limit / offset) + 1
	}

	prevPageNumber := resultsPageNumber - 1
	if prevPageNumber < 0 {
		prevPageNumber = 0
	}

	atLastPage := ((resultsPageNumber-1)*limit)+len(pages) >= count
	collection := PageCollection{
		pages,
		count,
		resultsPageNumber,
		limit,
		resultsPageNumber + 1,
		resultsPageNumber - 1,
		atLastPage,
	}
	return &collection, nil
}

// CountAllPages returns the number of page records
func CountAllPages() (int, error) {
	sql := `SELECT COUNT(*) FROM pages;`

	var count int

	err := DB.QueryRow(context.Background(), sql).Scan(&count)
	if err != nil {
		return 0, err
	}

	return count, nil
}
