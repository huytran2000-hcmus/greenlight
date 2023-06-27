include .envrc
# ==================================================================================== #
# HELPERS
# ==================================================================================== #
## help: print this help message
.PHONY: help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' | sed -e 's/^/ /'

## run/api: run the cmd/api application
.PHONY: run/api
run/api:
	echo $$GREENLIGHT_DB_DSN
	go run ./cmd/api -dsn $$GREENLIGHT_DB_DSN

# ==================================================================================== #
# DEVELOPMENT
# ==================================================================================== #
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
# ==================================================================================== #
# QUALITY CONTROL
# ==================================================================================== #
.PHONY: audit
audit:
	@echo 'Tidying and verifying module dependencies...'
	go mod tidy
	go mod verify
	@echo 'Formating code'
	gofumpt -l -w .
	goimports-reviser -set-alias ./...
	@echo 'Vetting code...'
	go vet ./...
	staticcheck ./...
	@echo 'Runnings tests'
	go test -race -vet=off ./...
