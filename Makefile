include .env

dev:
	./scripts/dev.sh

test:
	./scripts/test.sh

migrate-up:
	./scripts/migrate-up.sh

goose_status:
	goose -dir migrations postgres ${DATABASE_URL} status	

goose_up:
	goose -dir migrations postgres ${DATABASE_URL} up	

goose_down:
	goose -dir migrations postgres ${DATABASE_URL} down	

goose_reset:
	goose -dir migrations postgres ${DATABASE_URL} reset	

print_db:
	@echo "$(DATABASE_URL)"

compose_down:
	docker compose down

psql_db:
	docker compose exec db psql -U postgres -d internship_tracker
