# simplebank

## Requirements

1. Go 1.22
2. Docker
3. [golang-migrate](https://github.com/golang-migrate/migrate)
4. [sqlc](https://github.com/sqlc-dev/sqlc)
5. [mockgen](https://github.com/uber-go/mock)
6. [gRPC dependencies for go](https://grpc.io/docs/languages/go/quickstart/)
7. [statik](https://github.com/rakyll/statik)

Database can be set up via Makefile

## Useful commands

```shell
# Create a new SQL migration under db/migration directory 
migrate create -ext sql -dir db/migration -seq <filename>
```