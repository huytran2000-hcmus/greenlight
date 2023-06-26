## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	go run ./cmd/api

## db/psql: enter a psql repl connect to database
.PHONY: db/psql
db/psql:
	psql ${GREENLIGHT_DB_DSN}

## db/migrations/up: apply all the database migrations
.PHONY: db/migrations/up
db/migrations/up: confirm
	@echo 'Running up migrations...'
	migrate --path ./migrations -database ${GREENLIGHT_DB_DSN} up

.PHONY: db/migrations/new
## db/migrations/new name=$1: create a new database migrations
db/migrations/new:
	@echo 'Create migration files for ${name}...'
	migrate create -seq -ext=sql -dir=./migrations ${name}

confirm:
	@echo -n 'Are you sure? [y/N]' && read ans && [ $${ans:-N} = y]
