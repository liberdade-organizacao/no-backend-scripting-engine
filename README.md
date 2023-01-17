# Scripting Engine for Backend-as-a-Service

Add-on for [Liberdade's BaaS](https://github.com/liberdade-organizacao/no-backend-api)

## Quickstart

Requirements:
- Go
- PostgreSQL
- Docker (for development)
- BaaS API

From the BaaS API folder, setup the database and run the migrations.

To build and run the scripting engine executable:

``` sh
go build -o main.exe main/main.go
./main.exe up
```

The main execution can configured using the following environment variables:

| Variable name | Default value |
|---------------|---------------|
| `SERVER_PORT` | ":7781"       |
| `JDBC_DATABASE_URL` | connects to the Docker PostgreSQL database |

### Compilation Notes

It might be required to compile this project for different architectures.
In this case, cross-compilation should be available from Go's compiler:

``` sh
env GOOS=target-OS GOARCH=target-architecture go build package-import-path
```

For example:

``` sh
env GOOS=linux GOARCH=arm go build -o main.exe main/main.go
```

## Development

To run unit tests:

``` sh
make test
```


