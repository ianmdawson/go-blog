project_files = main.go

ifdef DATABASE_URL
	DATABASE_URL := $(DATABASE_URL)
else
	DATABASE_URL := 'postgres://goblog:password@localhost:5432'
endif

DEV_DATABASE_URL := $(DATABASE_URL)/blog_dev?sslmode=disable
TEST_DATABASE_URL := $(DATABASE_URL)/blog_test?sslmode=disable

test:
	go test ./...

build-run:
	go build $(project_files)
	./main

build:
	go build $(project_files)

docker-shell:
	docker-compose run blog /bin/ash

docker-up:
	docker-compose up --build

dependencies:
	go mod download
	go get -u github.com/pressly/goose/cmd/goose

db-setup:
	psql $(DATABASE_URL)?sslmode=disable -c "CREATE DATABASE blog_dev;"
	psql $(DATABASE_URL)?sslmode=disable -c "CREATE DATABASE blog_test;"

db-drop:
	psql $(DATABASE_URL)?sslmode=disable -c "DROP DATABASE IF EXISTS blog_dev;"
	psql $(DATABASE_URL)?sslmode=disable -c "DROP DATABASE IF EXISTS blog_test;"

reset-db: db-drop db-setup migrate

migrate:
	goose postgres $(DEV_DATABASE_URL) up
	goose postgres $(TEST_DATABASE_URL) up

migrate-status:
	goose postgres $(DEV_DATABASE_URL) status
	goose postgres $(TEST_DATABASE_URL) status

migrate-down:
	goose postgres $(DEV_DATABASE_URL) status
	goose postgres $(TEST_DATABASE_URL) status

migrate-reset:
	goose postgres $(DEV_DATABASE_URL) reset
	goose postgres $(TEST_DATABASE_URL) reset
