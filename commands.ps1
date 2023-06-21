function postgres() {
    docker run --name postgres12 -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:latest
}

function createdb() {
    docker exec -it postgres12 createdb --username=root --owner=root simple_bank
}

function dropdb() {
    docker exec -it postgres12 dropdb simple_bank
}

function addmigration {
    param(
        [Parameter(Mandatory=$true)]
        [string]$Name
    )

    migrate create -ext sql -dir db/migration -format "20060102150405" $Name
}

function migrateup () {
    migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose up
    regenerateschema

}

function migratedown () {
    param(
        [Parameter(Mandatory=$false)]
        [string]$Version 
    )   

    migrate -path db/migration -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" -verbose down $Version
    regenerateschema
}

function sqlcgen () {
    docker run --rm -v ".:/src" -w /src kjconroy/sqlc generate
}

function test {
    go test -v -cover ./...
}

function lscommands {
    echo "postgres`ncreatedb`ndropdb`naddmigration`nmigrateup`nmigratedown`nsqlcgen`ntest"
}

function regenerateschema {
    # Ejecutar pg_dump dentro del contenedor Docker y escribir el resultado en schema_aux.sql
    docker exec postgres12 pg_dump --schema-only --no-owner simple_bank > ./db/schema.sql

    # Leer el contenido del archivo
    $content = Get-Content -Encoding Unicode "./db/schema.sql"

    # Crear un objeto de codificación UTF-8 sin BOM
    $utf8NoBom = New-Object System.Text.UTF8Encoding $false

    $absolutePath = Resolve-Path "./db/schema.sql"

    # Escribir el contenido en el archivo con la codificación UTF-8 sin BOM
    [System.IO.File]::WriteAllLines($absolutePath, $content, $utf8NoBom)

}