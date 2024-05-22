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
	rm -f doc/swagger/*.swagger.json
	protoc --go-grpc_out=core/pb --go_out=core/pb --proto_path=proto --go-grpc_opt=paths=source_relative \
	--go_opt=paths=source_relative --grpc-gateway_out=core/pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=thread-generator \
	./proto/*.proto


.PHONY: build-web-proto
build-web-proto:
	rm -rf web/grpc/*
	protoc --proto_path ./proto \
		--plugin=protoc-gen-grpc-web=web/node_modules/.bin/protoc-gen-grpc-web \
		--plugin=protoc-gen-js=web/node_modules/.bin/protoc-gen-js \
		--js_out=import_style=commonjs,binary:web/grpc \
		--grpc-web_out=import_style=typescript,mode=grpcwebtext:web/grpc \
		 `find ./proto -name '*.proto'`

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
	cp .env_sample .env && \
	make generate-sim-key && \
	make generate-nextauth-secret
