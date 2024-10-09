dbdriver:
	export GOOSE_DRIVER=postgres

dbstring:
	export GOOSE_DBSTRING='postgres://huuthong:postgres@localhost:5432/bookmark'
migrate-up: dbdriver dbstring
	@export GOOSE_DRIVER=postgres && \
	export GOOSE_DBSTRING='postgres://huuthong:postgres@localhost:5432/bookmark' && \
	goose up
migrate-down: dbdriver dbstring
	@export GOOSE_DRIVER=postgres && \
	export GOOSE_DBSTRING='postgres://huuthong:postgres@localhost:5432/bookmark' && \
	goose down
