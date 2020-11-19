package page

import (
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
