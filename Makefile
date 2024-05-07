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
	rm -f pkg/pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --go-grpc_out=pkg/pb --go_out=pkg/pb --proto_path=proto --go-grpc_opt=paths=source_relative \
	--go_opt=paths=source_relative --grpc-gateway_out=pkg/pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=thread-generator \
	./proto/*.proto


.PHONY: build-web-proto
build-web-proto:
	rm -rf web/src/pb/*
	protoc --proto_path ./proto \
		--plugin=protoc-gen-grpc-web=web/node_modules/.bin/protoc-gen-grpc-web \
		--plugin=protoc-gen-ts_proto=web/node_modules/.bin/protoc-gen-ts_proto \
		--ts_out=import_style=commonjs,binary:web/src/pb \
		--grpc-web_out=import_style=typescript,mode=grpcwebtext:web/src/pb \
		--ts_proto_out=./web/src/pb \
		--ts_proto_opt=env=browser \
		--ts_proto_opt=useOptionals=true \
		--ts_proto_opt=unrecognizedEnum=false \
		 `find ./proto -name '*.proto'`

.PHONY: generate-sim-key
generate-sim-key:
	TOKEN_SYMMETRIC_KEY=$$(openssl rand -hex 16); \
	echo "TOKEN_SYMMETRIC_KEY=$$TOKEN_SYMMETRIC_KEY" >> .env
