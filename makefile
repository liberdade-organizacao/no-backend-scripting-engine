.PHONY: default
default: run

.PHONY: psql
psql:
	psql -h localhost -p 5434 -d baas -U liberdade -W

.PHONY: test
test:
	go test ./database
	go test ./services/*.go

.PHONY: build
build: test
	go build -o main.exe main/main.go

.PHONY: run
run: build
	./main.exe

.PHONY: migrate_up
migrate_up: build
	./main.exe migrate_up

.PHONY: migrate_down
migrate_down: build
	./main.exe migrate_down


