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

.PHONY: proto
proto:
	@echo "🔧 Generating protocol buffers..."
	@echo "Checking for required tools..."
	@test -f "$(shell go env GOPATH)/bin/protoc-gen-go" || (echo "❌ protoc-gen-go not found. Installing..." && go install google.golang.org/protobuf/cmd/protoc-gen-go@latest)
	@test -f "$(shell go env GOPATH)/bin/protoc-gen-connect-go" || (echo "❌ protoc-gen-connect-go not found. Installing..." && go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest)
	@test -f "$(shell go env GOPATH)/bin/protoc-gen-openapiv2" || (echo "❌ protoc-gen-openapiv2 not found. Installing..." && go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest)
	@echo "✅ All tools available"
	@echo "Creating output directories..."
	@mkdir -p core/pb/pbconnect
	@mkdir -p api/openapi
	@echo "Generating Go and Connect-RPC files..."
	@cd proto && PATH="$(shell go env GOPATH)/bin:$$PATH" buf generate --template buf.gen.make.yaml
	@echo "✅ Protocol buffers generated successfully!"
	@echo "📁 Generated files:"
	@echo "   - Go types: core/pb/"
	@echo "   - Connect-RPC: core/pb/pbconnect/"
	@echo "   - OpenAPI: api/openapi/"

.PHONY: proto-clean
proto-clean:
	@echo "🧹 Cleaning generated protocol buffer files..."
	@rm -rf core/pb/*.pb.go
	@rm -rf core/pb/pbconnect/
	@rm -rf api/openapi/
	@echo "✅ Protocol buffer files cleaned"

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

.PHONY: run-migrations
run-migrations:
	@echo "🔄 Running database migrations..."
	@docker-compose run --rm migrations || { echo "❌ Migration failed"; exit 1; }
	@echo "✅ Database migrations completed successfully"

.PHONY: generate-models
generate-models:
	@echo "🔄 Generating database models using SQLBoiler..."
	@docker-compose run --rm generate-models || { echo "❌ Model generation failed"; exit 1; }
	@echo "✅ Database models generated successfully"
	@echo "📁 Generated models in: core/db/models/"

# This rule allows capturing arbitrary targets as arguments
%:
	@:
