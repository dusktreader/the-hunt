.ONESHELL:
.DEFAULT_GOAL:=help
SHELL:=/bin/bash
DB_URL=postgres://compose-db-user:compose-db-pswd@db:5432/compose-db-name?sslmode=disable


# ==== Helpers =========================================================================================================
.PHONY: confirm
confirm:  ## Don't use this directly
	@echo -n "Are you sure? [y/N] " && read ans && [ $${ans:-N} = y ]


.PHONY: help
help:  ## Show help message
	@awk 'BEGIN {FS = ": .*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% 0-9a-zA-Z_\/-]+(\\:[$$()% 0-9a-zA-Z_\/-]+)*:.*?##/ { gsub(/\\:/,":", $$1); printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)


.PHONY: clean
clean:  ## Clean up build artifacts and other junk
	@rm -rf bin/*


# ==== Database Commands ===============================================================================================
.PHONY: db/login
db/login:  ## Log into the database (in docker compose)
	@docker compose run \
    	--rm \
    	db \
    	psql ${DB_URL}


.PHONY: db/one-up
db/one-up:  ## Apply one upward migration to the database (in docker compose)
	@docker compose run --rm migrate goose up-by-one


.PHONY: db/up
db/up:  ## Apply all upward migrations to the database (in docker compose)
	@docker compose run --rm migrate goose up


.PHONY: db/one-up
db/one-down:  ## Apply one downward migration to the database (in docker compose)
	@docker compose run --rm migrate goose down-by-one


.PHONY: db/down
db/down:  ## Apply all downward migrations to the database (in docker compose)
	@docker compose run --rm migrate goose down


.PHONY: db/migrate
db/migrate: name ?= "unnamed_migration"
db/migrate:  ## Create a new database migration (set name with name=<migration_name>)
	@docker compose run \
    	--rm \
    	--no-deps \
    	--volume=/etc/passwd:/etc/passwd:ro \
    	--volume=/etc/group:/etc/group:ro \
    	--user="$(shell id -u):$(shell id -g)" \
    	migrate goose create -s ${name} sql && \
	${EDITOR} migrations/*${name}.sql


.PHONY: db/nuke
db/nuke: confirm  ## Nuke the database (in docker compose)
	@docker compose run \
    	--rm \
    	--volume=$(pwd)/etc/postgres/nuke.sql:/nuke.sql:ro \
    	db \
    	psql ${DB_URL} \
    	-f /nuke.sql


# ==== Quality Control =================================================================================================
.PHONY: qa/tidy
qa/tidy:  ## Clean up dependencies and format source files
	go mod tidy
	go fmt ./...


.PHONY: qa/audit
qa/audit:
	go mod tidy -diff
	go mod verify
	go vet ./...
	go tool staticcheck ./...
	go test -race -vet=off ./...


# ==== App Commands ====================================================================================================
.PHONY: run
app/run:  ## Run the API
	go run ./cmd/api


.PHONY: build
app/build: ldflags ?= '-s'
app/build: GOOS_PART := $(if $(GOOS),.$(GOOS),)
app/build: GOARCH_PART := $(if $(GOARCH),.$(GOARCH),)
app/build: OUT_FILE :=./bin/the-hunt-api${GOOS_PART}
app/build:  ## Build the API executable
	GOOS=${GOOS} GOARCH=${GOARCH} go build -ldflags=${ldflags} -o=${OUT_FILE} ./cmd/api
