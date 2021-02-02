FROM golang:1.15-alpine

# The latest alpine images don't have some tools like (`git` and `bash`).
# Adding git, bash and openssh to the image
RUN apk update && apk upgrade && \
  apk add --no-cache bash git openssh \
  make postgresql-client \
  gcc musl-dev

RUN go get -u github.com/pressly/goose/cmd/goose

WORKDIR /src/ianmdawson/go-blog
# TODO:
# COPY go.mod go.sum ./
# RUN go mod download

COPY . .

RUN go build -o main .

EXPOSE 8080

CMD ["/src/ianmdawson/go-blog/main"]
