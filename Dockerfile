FROM golang:1.21.1-alpine3.18

WORKDIR /usr/src/app
COPY go.mod go.sum ./
RUN go mod download && go mod verify

COPY . .
RUN go build -v -o /usr/local/bin/app ./main/main.go

CMD ["app", "up"]
