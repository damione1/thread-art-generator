.PHONY: up
up:
	tilt up

.PHONY: down
down:
	tilt down

.PHONY: restart
restart:
	tilt down && tilt up


.PHONY: generate-sim-key
generate-sim-key:
	TOKEN_SYMMETRIC_KEY=$$(openssl rand -hex 16); \
	echo "TOKEN_SYMMETRIC_KEY=$$TOKEN_SYMMETRIC_KEY" >> .env

.PHONY: generate-nextauth-secret
generate-nextauth-secret:
	NEXTAUTH_SECRET=$$(openssl rand -hex 32); \
	echo "NEXTAUTH_SECRET=$$NEXTAUTH_SECRET" >> .env

.PHONY: init_env
init_env:
	cp .env.sample .env && \
	make generate-sim-key && \
	make generate-nextauth-secret

.PHONY: setup
setup: init_env
	@echo "\n\n======================="
	@echo "Initial setup complete!"
	@echo "======================="
	@echo "\nEnvironment variables set in .env"
	@echo "Run 'make proto' to generate proto files"

.PHONY: psql
psql:
	@eval $$(grep -e "POSTGRES_USER\|POSTGRES_PASSWORD\|POSTGRES_DB" .env | sed 's/^/export /'); \
	PGPASSWORD=$$POSTGRES_PASSWORD psql -h localhost -U $$POSTGRES_USER -d $$POSTGRES_DB
