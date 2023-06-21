.PHONY: postgres createdb dropdb addmigration migrateup migratedown sqlcgen test lscommands regenerateschema

postgres:
	docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest

createdb:
	docker exec -it postgres12 createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it postgres12 dropdb simple_bank

addmigration:
	@if [ "${name}" = "" ]; then \
		echo "Error: Debes especificar el nombre de la migración con name=<nombre_de_la_migración>"; \
		exit 1; \
	fi
	migrate create -ext sql -dir db/migration -format "20060102150405" $(name)

migrateup:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down $(version)

sqlcgen:
	docker run --rm -v ${PWD}:/src -w /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

lscommands:
	@echo "postgres\ncreatedb\ndropdb\naddmigration\nmigrateup\nmigratedown\nsqlcgen\ntest"

regenerateschema:
	docker exec postgres12 pg_dump --schema-only --no-owner simple_bank > ./db/schema.sql
	iconv -f UTF-16 -t UTF-8 ./db/schema.sql > ./db/schema_aux.sql
	mv ./db/schema_aux.sql ./db/schema.sql
