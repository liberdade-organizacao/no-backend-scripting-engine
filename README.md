# Scripting Engine for Backend-as-a-Service

> WORK IN PROGRESS!

Add-on for [Liberdade's BaaS](https://github.com/liberdade-organizacao/no-backend-api)

## Setup

Requirements:
- Go
- PostgreSQL
- Docker (for development)
- BaaS API

From the BaaS API folder, setup the database and run the migrations.

To build and run the scripting engine executable:

``` sh
go build -o main.exe main/main.go
./main.exe
```

## Development

To run unit tests:

``` sh
make test
```


