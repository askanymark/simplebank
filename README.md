# simplebank

## Requirements

1. Go 1.25.5
2. Docker
3. [golang-migrate](https://github.com/golang-migrate/migrate)
4. [sqlc](https://github.com/sqlc-dev/sqlc)
5. [mockgen](https://github.com/uber-go/mock)
6. [gRPC dependencies for go](https://grpc.io/docs/languages/go/quickstart/)
7. [statik](https://github.com/rakyll/statik)

Database can be set up via Makefile

## Things to know

- There are two APIs available. gRPC is the main one, and the REST is available via HTTP gateway provided by `grpc-gateway`
  - At this moment, api package written in Gin is not used and will be removed
- `sqlc` generates Go code for the database from SQL schema
- `mockgen` generates mocks for interfaces that cover the database and task distributor 
- OpenAPI specification is available in the form of Swagger UI at http://localhost:8080/swagger. It is generated from the gRPC service definition inside `proto` directory. The static assets then get compiled into serveable Go file using `statik`

## Useful commands

```shell
# Create a new SQL migration under db/migration directory 
migrate create -ext sql -dir db/migration -seq <filename>
```