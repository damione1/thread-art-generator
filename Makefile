.PHONY: help
help:
	@echo "Thread Art Generator - Development Commands"
	@echo ""
	@echo "🚀 Development Environment:"
	@echo "  make setup          - Run initial local development setup"
	@echo "  make up             - Start all services with Tilt"
	@echo "  make down           - Stop all services"
	@echo "  make restart        - Restart all services"
	@echo ""
	@echo "🔐 Security & Keys:"
	@echo "  make generate-keys  - Generate secure 32-byte keys for PASETO/cookies"
	@echo "  make update-env-keys - Update .env file with new secure keys"
	@echo "  make validate-keys  - Validate that keys are properly formatted"
	@echo ""
	@echo "🔧 Code Generation:"
	@echo "  make proto          - Generate protocol buffer files"
	@echo "  make proto-clean    - Clean generated protocol buffer files"
	@echo "  make generate-models - Generate SQLBoiler models from database"
	@echo "  make generate-templ - Generate Templ templates"
	@echo ""
	@echo "🗄️  Database:"
	@echo "  make psql              - Connect to PostgreSQL database"
	@echo "  make migration <name>  - Create new database migration"
	@echo "  make run-migrations    - Run all pending database migrations"
	@echo ""
	@echo "🔥 Firebase:"
	@echo "  make firebase-build - Build Firebase functions"
	@echo "  make firebase-start - Start Firebase emulator suite"

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
	@echo "Installing golang-migrate if needed..."
	@test -f "$(shell go env GOPATH)/bin/migrate" || go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest
	@echo "Running migrations..."
	@eval $$(grep -E "POSTGRES_(USER|PASSWORD|DB)" .env | sed 's/^/export /'); \
	$(shell go env GOPATH)/bin/migrate -path core/db/migrations -database "postgres://$$POSTGRES_USER:$$POSTGRES_PASSWORD@localhost:5432/$$POSTGRES_DB?sslmode=disable" up || { echo "❌ Migration failed"; exit 1; }
	@echo "✅ Database migrations completed successfully"

.PHONY: generate-models
generate-models:
	@echo "🔄 Generating database models using SQLBoiler..."
	@echo "Installing SQLBoiler if needed..."
	@test -f "$(shell go env GOPATH)/bin/sqlboiler" || go install github.com/volatiletech/sqlboiler/v4@v4.16.2
	@test -f "$(shell go env GOPATH)/bin/sqlboiler-psql" || go install github.com/volatiletech/sqlboiler/v4/drivers/sqlboiler-psql@v4.16.2
	@echo "Generating models..."
	@eval $$(grep -E "POSTGRES_(USER|PASSWORD|DB)" .env | sed 's/^/export /'); \
	PSQL_HOST=localhost PSQL_PORT=5432 PSQL_DBNAME=$$POSTGRES_DB PSQL_USER=$$POSTGRES_USER PSQL_PASS=$$POSTGRES_PASSWORD PSQL_SSLMODE=disable \
	PATH="$(shell go env GOPATH)/bin:$$PATH" $(shell go env GOPATH)/bin/sqlboiler psql || { echo "❌ Model generation failed"; exit 1; }
	@echo "✅ Database models generated successfully"
	@echo "📁 Generated models in: core/db/models/"

.PHONY: generate-templ
generate-templ:
	@echo "🔄 Building npm packages..."
	@cd client && npm install && npm run build
	@echo "✅ Npm packages built successfully"
	@echo "🔄 Generating Templ templates..."
	@echo "Installing Templ if needed..."
	@test -f "$(shell go env GOPATH)/bin/templ" || go install github.com/a-h/templ/cmd/templ@latest
	@echo "Generating templates..."
	@PATH="$(shell go env GOPATH)/bin:$$PATH" $(shell go env GOPATH)/bin/templ generate || { echo "❌ Template generation failed"; exit 1; }
	@echo "✅ Templ templates generated successfully"

.PHONY: firebase-build
firebase-build:
	@echo "🔧 Building Firebase Functions..."
	@cd functions && npm install && npm run build
	@echo "✅ Firebase Functions built successfully"

.PHONY: firebase-start
firebase-start: firebase-build
	@echo "🚀 Starting Firebase Emulator Suite..."
	@PATH="/opt/homebrew/opt/openjdk/bin:$$PATH" firebase emulators:start --only auth,functions,pubsub,storage,ui --project demo-thread-art-generator

.PHONY: generate-keys
generate-keys:
	@echo "🔐 Generating secure keys for PASETO authentication..."
	@echo "PASETO_SECRET_KEY=$$(openssl rand -hex 32 | head -c 32)"
	@echo "COOKIE_HASH_KEY=$$(openssl rand -hex 32 | head -c 32)"
	@echo "COOKIE_BLOCK_KEY=$$(openssl rand -hex 32 | head -c 32)"
	@echo ""
	@echo "🔧 Copy these keys to your .env file:"
	@echo "   PASETO_SECRET_KEY=$$(openssl rand -hex 32 | head -c 32)"
	@echo "   COOKIE_HASH_KEY=$$(openssl rand -hex 32 | head -c 32)"
	@echo "   COOKIE_BLOCK_KEY=$$(openssl rand -hex 32 | head -c 32)"
	@echo ""
	@echo "⚠️  IMPORTANT: Keep these keys secure and consistent across deployments!"

.PHONY: update-env-keys
update-env-keys:
	@echo "🔄 Updating .env file with new secure keys..."
	@if [ ! -f .env ]; then echo "❌ .env file not found. Run 'make setup' first."; exit 1; fi
	@PASETO_KEY=$$(openssl rand -hex 32 | head -c 32); \
	HASH_KEY=$$(openssl rand -hex 32 | head -c 32); \
	BLOCK_KEY=$$(openssl rand -hex 32 | head -c 32); \
	sed -i.bak \
		-e "s/^PASETO_SECRET_KEY=.*/PASETO_SECRET_KEY=$$PASETO_KEY/g" \
		-e "s/^COOKIE_HASH_KEY=.*/COOKIE_HASH_KEY=$$HASH_KEY/g" \
		-e "s/^COOKIE_BLOCK_KEY=.*/COOKIE_BLOCK_KEY=$$BLOCK_KEY/g" \
		.env
	@echo "✅ Updated .env file with new secure keys"
	@echo "🔒 Backup saved as .env.bak"
	@echo ""
	@echo "Generated keys:"
	@grep -E "^(PASETO_SECRET_KEY|COOKIE_HASH_KEY|COOKIE_BLOCK_KEY)=" .env

.PHONY: validate-keys
validate-keys:
	@echo "🔍 Validating PASETO and cookie keys..."
	@if [ ! -f .env ]; then echo "❌ .env file not found. Run 'make setup' first."; exit 1; fi
	@PASETO_KEY=$$(grep "^PASETO_SECRET_KEY=" .env | cut -d'=' -f2); \
	HASH_KEY=$$(grep "^COOKIE_HASH_KEY=" .env | cut -d'=' -f2); \
	BLOCK_KEY=$$(grep "^COOKIE_BLOCK_KEY=" .env | cut -d'=' -f2); \
	if [ -z "$$PASETO_KEY" ]; then echo "❌ PASETO_SECRET_KEY not found in .env"; exit 1; fi; \
	if [ $${#PASETO_KEY} -ne 32 ]; then echo "❌ PASETO_SECRET_KEY must be exactly 32 bytes, got $${#PASETO_KEY}"; exit 1; fi; \
	if [ -z "$$HASH_KEY" ]; then echo "❌ COOKIE_HASH_KEY not found in .env"; exit 1; fi; \
	if [ $${#HASH_KEY} -ne 32 ]; then echo "❌ COOKIE_HASH_KEY must be exactly 32 bytes, got $${#HASH_KEY}"; exit 1; fi; \
	if [ -z "$$BLOCK_KEY" ]; then echo "❌ COOKIE_BLOCK_KEY not found in .env"; exit 1; fi; \
	if [ $${#BLOCK_KEY} -ne 32 ]; then echo "❌ COOKIE_BLOCK_KEY must be exactly 32 bytes, got $${#BLOCK_KEY}"; exit 1; fi; \
	echo "✅ All keys are properly formatted (32 bytes each)"; \
	echo "✅ PASETO_SECRET_KEY: $$PASETO_KEY"; \
	echo "✅ COOKIE_HASH_KEY: $$HASH_KEY"; \
	echo "✅ COOKIE_BLOCK_KEY: $$BLOCK_KEY"

# This rule allows capturing arbitrary targets as arguments
%:
	@:
