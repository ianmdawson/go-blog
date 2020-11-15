project_files = wiki.go

build-run:
	go build $(project_files)
	./wiki

build:
	go build $(project_files)
