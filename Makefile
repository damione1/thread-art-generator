.PHONY: up
up:
	tilt up

.PHONY: down
down:
	tilt down

.PHONY: restart
restart:
	tilt down && tilt up



.PHONY: build-go-proto
build-go-proto:
	rm -f core/pb/*.go
	protoc --go-grpc_out=core/pb --go_out=core/pb --proto_path=proto --go-grpc_opt=paths=source_relative \
	--go_opt=paths=source_relative --grpc-gateway_out=core/pb --grpc-gateway_opt=paths=source_relative \
	--govalidators_out=paths=source_relative:core/pb \
	./proto/*.proto

.PHONY: build-web-proto
build-web-proto:
	@echo "Generating NextJS gRPC client code..."
	docker-compose run --rm ts-proto-generator
	@echo "Web client code generation complete!"

.PHONY: install-web-proto-deps
install-web-proto-deps:
	@echo "Installing web proto dependencies..."
	cd web && \
	npm install --legacy-peer-deps --save-dev \
		@bufbuild/buf \
		@bufbuild/protoc-gen-es \
		@connectrpc/protoc-gen-connect-es && \
	npm install --legacy-peer-deps --save \
		@bufbuild/connect \
		@bufbuild/connect-web \
		@bufbuild/protobuf
	@echo "Web proto dependencies installed!"


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
setup: init_env install-go-proto-deps
	@echo "\n\n======================="
	@echo "Initial setup complete!"
	@echo "======================="
	@echo "\nEnvironment variables set in .env"
	@echo "Go proto files generated in core/pb/"
