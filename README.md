# How to run 
1. create .env with
```sh
CSV_IMPORTER_DB_HOST=localhost
CSV_IMPORTER_DB_PORT=5432
CSV_IMPORTER_DB_USER=postgres
CSV_IMPORTER_DB_PASSWORD=mypassword
CSV_IMPORTER_DB_NAME=postgres
```
2. run docker compose 
```sh
docker compose up -d
```

3. source .env
```sh
export $(grep -v '^#' .env | xargs)
```

4. run
```sh
go run ./cmd/csv-importer
```
