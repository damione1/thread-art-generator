.PHONY: up
up:
	tilt up

.PHONY: down
down:
	tilt down

.PHONY: restart
restart:
	tilt down && tilt up

.PHONY: setup
setup:
	@echo "Running local development setup..."
	@./scripts/local_setup.sh

.PHONY: psql
psql:
	@eval $$(grep -e "POSTGRES_USER\|POSTGRES_PASSWORD\|POSTGRES_DB" .env | sed 's/^/export /'); \
	PGPASSWORD=$$POSTGRES_PASSWORD psql -h localhost -U $$POSTGRES_USER -d $$POSTGRES_DB

.PHONY: migration
migration:
	@if [ -z "$(filter-out $@,$(MAKECMDGOALS))" ]; then \
		echo "Error: Migration name required. Usage: make migration Your Migration Name"; \
		exit 1; \
	fi; \
	migration_name=$$(echo "$(filter-out $@,$(MAKECMDGOALS))" | tr '[:upper:]' '[:lower:]' | tr ' ' '_'); \
	latest_num=$$(ls -1 core/db/migrations/*.up.sql 2>/dev/null | sed 's/.*\/\([0-9]\{6\}\)_.*/\1/' | sort -nr | head -1 || echo "000000"); \
	next_num=$$(printf "%06d" $$((10#$${latest_num} + 1))); \
	echo "Creating migration $$next_num"_"$$migration_name"; \
	touch "core/db/migrations/$$next_num"_"$$migration_name.up.sql"; \
	touch "core/db/migrations/$$next_num"_"$$migration_name.down.sql"; \
	echo "-- Migration $$next_num: $$migration_name (up)" > "core/db/migrations/$$next_num"_"$$migration_name.up.sql"; \
	echo "-- Migration $$next_num: $$migration_name (down)" > "core/db/migrations/$$next_num"_"$$migration_name.down.sql"; \
	echo "Created migration files:"; \
	echo "  - core/db/migrations/$$next_num"_"$$migration_name.up.sql"; \
	echo "  - core/db/migrations/$$next_num"_"$$migration_name.down.sql"

# This rule allows capturing arbitrary targets as arguments
%:
	@:
