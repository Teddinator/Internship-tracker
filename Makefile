include .env

dev:
	./scripts/dev.sh

test:
	./scripts/test.sh

migrate-up:
	./scripts/migrate-up.sh

migrate-down:
	./scripts/migrate-down.sh

seed:
	./scripts/seed.sh

backup-db:
	./scripts/backup-db.sh

