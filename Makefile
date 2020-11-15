project_files = main.go

build-run:
	go build $(project_files)
	./main

build:
	go build $(project_files)

docker-shell:
	docker-compose run app /bin/ash

docker-up:
	docker-compose up --build