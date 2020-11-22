package models

import (
	"github.com/jackc/pgx/v4"
)

// DB holds the database connection for all models
var DB *pgx.Conn
