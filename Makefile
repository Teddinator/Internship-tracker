include .env

goose_status:
	goose -dir migrations postgres ${DATABASE_URL} status	

goose_up:
	goose -dir migrations postgres ${DATABASE_URL} up	

goose_down:
	goose -dir migrations postgres ${DATABASE_URL} down	

print_db:
	@echo "$(DATABASE_URL)"

compose_up:
	docker compose up

compose_down:
	docker compose down