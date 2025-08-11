postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:12-alpine

createdb:
	docker exec -it postgres createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

migrateup:
	goose -dir sql/schema postgres "postgres://postgres:1234@localhost:5432/simple_bank?sslmode=disable" up

migratedown:
	goose -dir sql/schema postgres "postgres://postgres:1234@localhost:5432/simple_bank?sslmode=disable" down

generatesql:
	sqlc generate
test:
	go test -v -cover ./...

server:
	go run main.go
mock:
	mockgen -destination internal/database/mock/store.go github.com/Glenn444/banking-app/internal/database Store



.PHONY: postgres createdb dropdb migrateup migratedown generatesql test server mock