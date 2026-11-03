.PHONY: default
default: run

.PHONY: test
test:
	go test ./common/*.go
	go test ./database/*.go
	go test ./controllers/*.go

.PHONY: build
build: test
	go build -ldflags "-w" -o main.exe main/main.go

.PHONY: run
run: build
	./main.exe up

.PHONY: docker-build
docker-build:
	docker build -t scripting-engine .

.PHONY: docker-run
docker-run: docker-build
	docker run -p 127.0.0.1:7781:7781 scripting-engine

