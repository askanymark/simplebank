# simplebank

## Requirements

1. Go 1.24.1
2. Docker
3. [golang-migrate](https://github.com/golang-migrate/migrate)
4. [sqlc](https://github.com/sqlc-dev/sqlc)
5. [mockgen](https://github.com/uber-go/mock)
6. [gRPC dependencies for go](https://grpc.io/docs/languages/go/quickstart/)
7. [statik](https://github.com/rakyll/statik)

Database can be set up via Makefile

## Things to know

- Use `make sqlc` to generate Go code for the database from SQL schema
- mockgen is used to generate mocks for interfaces that cover the database and task distributor 
- There is an OpenAPI specification available in the form of Swagger UI at http://localhost:8080/swagger. It is generated from the gRPC service definition inside `proto` directory and then compiled into static assets using statik. This is what HTTP gateway is used for

## Useful commands

```shell
# Create a new SQL migration under db/migration directory 
migrate create -ext sql -dir db/migration -seq <filename>
```