.PHONY: default
default: run

.PHONY: psql
psql:
	psql -h localhost -p 5434 -d baas -U liberdade -W

.PHONY: test
test:
	go test ./common/*.go
	go test ./database/*.go
	go test ./controllers/*.go

.PHONY: build
build: test
	go build -o main.exe main/main.go

.PHONY: run
run: build
	./main.exe up

.PHONY: docker-build
docker-build:
	docker build -t scripting-engine .

.PHONY: docker-run
docker-run: docker-build
	docker run -p 127.0.0.1:7781:7781 scripting-engine

