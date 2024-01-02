.PHONY: up
up:
	tilt up

.PHONY: down
down:
	tilt down

.PHONY: restart
restart:
	tilt down && tilt up


.PHONY: build-proto
build-proto:
	rm -f pkg/pb/*.go
	rm -f doc/swagger/*.swagger.json
	protoc --go-grpc_out=pkg/pb --go_out=pkg/pb --proto_path=proto --go-grpc_opt=paths=source_relative \
	--go_opt=paths=source_relative --grpc-gateway_out=pkg/pb --grpc-gateway_opt=paths=source_relative \
	--openapiv2_out=doc/swagger --openapiv2_opt=allow_merge=true,merge_file_name=cookbook \
	./proto/*.proto
