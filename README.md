# go-blog

A personal project to get more familiar with go template rendering and a dockerized go server.

## Getting Started
Requirements: [Docker](https://www.docker.com/), [go](https://golang.org/) version 1.13 or higher.

Retrieve the source code
```
go get -u github.com/ianmdawson/go-blog
```

Build and run the application in the docker container.
```
cd ${GOPATH}/src/github.com/ianmdawson/go-blog
docker-compose up --build
```

### TODOs
- Better documentation
- Markdown handling
  - Allow posts to contain markdown for better looking posts.
- Test improvements:
  - Make database reset more efficient in tests
    - Use database/sql instead of jackc/pgx
    - Avoid requiring the database setup in tests via mocking: https:github.com/jackc/pgx/issues/616#issuecomment-535749087
  - More routing/http handler tests
- Users, authentication, and permissions:
  - Right now anyone can create or edit a post, it's almost like a free-for-all wiki. There's no user authentication or permissions yet.
