function postgres() {
    docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest
}

function createdb() {
    docker exec -it postgres12 createdb --username=root --owner=root simple_bank
}

function dropdb() {
    docker exec -it postgres12 dropdb simple_bank
}

function migrateup () {
    migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
}

function migratedown () {
    migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down
}

function sqlcgen () {
    docker run --rm -v ".:/src" -w /src kjconroy/sqlc generate
}