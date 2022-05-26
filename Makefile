postgres:
	docker run --name postgres14.3 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14.3-alpine

createdb:
	docker exec -it postgres14.3 createdb --username=root --owner=root twitter_wannabe

dropdb:
	docker exec -it postgres14.3 dropdb twitter_wannabe

migrateup:
	migrate -path database/migration -database "postgresql://root:secret@localhost:5432/twitter_wannabe?sslmode=disable" -verbose up

migratedown:
	migrate -path database/migration -database "postgresql://root:secret@localhost:5432/twitter_wannabe?sslmode=disable" -verbose down

sqlc:
	docker run --rm -v "D:\Study\learn_go\twitter_wannabe:/src" -w /src kjconroy/sqlc generate

.PHONY: postgres createdb dropdb migrateup migratedown sqlc