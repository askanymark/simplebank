# simplebank

## Requirements

1. Go 1.22
2. Docker
3. [golang-migrate](https://github.com/golang-migrate/migrate)

Database can be set up via Makefile

## Useful commands

```shell
# Create a new SQL migration under db/migration directory 
migrate create -ext sql -dir db/migration -seq <filename>
# Run migrations UP
migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
```