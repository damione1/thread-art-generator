.PHONY: up
up:
	tilt up

.PHONY: down
down:
	tilt down

.PHONY: restart
restart:
	tilt down && tilt up


.PHONY: install-go-proto-deps
install-go-proto-deps:
	@echo "Ensuring GOPATH/bin is in PATH..."
	@GOBIN=$$(go env GOPATH)/bin; \
	if [[ ":$$PATH:" != *":$$GOBIN:"* ]]; then \
		echo "Adding $$GOBIN to PATH for this session"; \
		export PATH=$$PATH:$$GOBIN; \
		echo "You may want to add the following to your shell profile:"; \
		echo "export PATH=\$$PATH:\$$(go env GOPATH)/bin"; \
	else \
		echo "GOPATH/bin is already in PATH"; \
	fi
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
	go install github.com/mwitkow/go-proto-validators/protoc-gen-govalidators@latest
	@echo "Verifying installations..."
	@which protoc-gen-go protoc-gen-go-grpc protoc-gen-grpc-gateway protoc-gen-openapiv2 protoc-gen-govalidators || \
		{ echo "Some tools are not in PATH. Please ensure GOPATH/bin is in your PATH."; \
		  echo "You can add it with: export PATH=\$$PATH:\$$(go env GOPATH)/bin"; \
		  exit 1; }


.PHONY: build-go-proto
build-go-proto:
	@GOBIN=$$(go env GOPATH)/bin; \
	if [[ ":$$PATH:" != *":$$GOBIN:"* ]]; then \
		echo "Adding $$GOBIN to PATH for this session"; \
		export PATH=$$PATH:$$GOBIN; \
	fi; \
	rm -f core/pb/*.go; \
	rm -f doc/swagger/*.swagger.json; \
	protoc --go-grpc_out=core/pb --go_out=core/pb --proto_path=proto --go-grpc_opt=paths=source_relative \
	--go_opt=paths=source_relative --grpc-gateway_out=core/pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=thread-generator \
	--govalidators_out=paths=source_relative:core/pb \
	./proto/*.proto


.PHONY: build-web-proto
build-web-proto:
	@echo "Generating NextJS gRPC client code..."
	rm -rf web/src/lib/api/*
	cd web && \
	npm install --save-dev \
		@bufbuild/buf \
		@bufbuild/protoc-gen-es \
		@connectrpc/protoc-gen-connect-es && \
	npm install --save \
		@bufbuild/connect \
		@bufbuild/connect-web \
		@bufbuild/protobuf && \
	mkdir -p src/lib/api && \
	npx protoc --es_out src/lib/api --es_opt target=ts \
		--connect-es_out src/lib/api --connect-es_opt target=ts \
		--proto_path=../proto \
		../proto/*.proto
	@echo "Web client code generation complete!"


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
