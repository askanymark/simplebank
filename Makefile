postgres:
	docker run --name postgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres dropdb simple_bank

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migrateup1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up 1

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down

migratedown1:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down 1

new_migration:
	migrate create -ext sql -dir db/migration -seq $(name)

sqlc:
	sqlc generate

redis:
	docker run --name redis -p 6379:6379 -d redis:alpine

test:
	GOTOOLCHAIN=go1.25.5+auto go test -v -cover ./...

server:
	go run main.go

mock:
	mockgen -package mockdb -destination db/mock/store.go simplebank/db/sqlc Store
	mockgen -package mockwk -destination worker/mock/distributor.go simplebank/worker TaskDistributor

proto:
	rm -rf pb/*.go
	rm -rf pb/accounts/*.go
	rm -rf pb/users/*.go
	rm -rf pb/transfers/*.go
	rm -f docs/swagger/*.swagger.json
	protoc \
		--proto_path=proto \
		--go_out=pb --go_opt=paths=source_relative \
		--go-grpc_out=pb --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=pb --grpc-gateway_opt=paths=source_relative \
		--openapiv2_out=docs/swagger \
		--openapiv2_opt=allow_merge=true,merge_file_name=simplebank \
		--experimental_allow_proto3_optional \
		proto/*.proto \
		proto/accounts/*.proto \
		proto/users/*.proto \
		proto/transfers/*.proto
	mv pb/accounts/*.go pb/ 2>/dev/null || true
	mv pb/users/*.go pb/ 2>/dev/null || true
	mv pb/transfers/*.go pb/ 2>/dev/null || true
	rm -rf pb/accounts pb/users pb/transfers
	statik -src=./docs/swagger -dest=./docs

.PHONY: postgres createdb migrateup sqlc mock proto redis