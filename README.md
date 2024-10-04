migrate-up:
	@export $(cat .env | xargs) && \
	goose up
