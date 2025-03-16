#!/bin/bash
set -e

echo "Generating Go proto files..."
rm -f /app/core/pb/*.go
protoc --go-grpc_out=/app/core/pb --go_out=/app/core/pb --proto_path=/app/proto --go-grpc_opt=paths=source_relative \
  --go_opt=paths=source_relative --grpc-gateway_out=/app/core/pb --grpc-gateway_opt=paths=source_relative \
  --govalidators_out=paths=source_relative:/app/core/pb \
  /app/proto/*.proto
echo "Go proto generation complete!"

echo "Generating TypeScript proto files..."
rm -rf /app/web/src/lib/pb/*
mkdir -p /app/web/src/lib/pb

# Find all .proto files in the /app/proto directory and its subdirectories
PROTO_FILES=$(find /app/proto -name "*.proto")

protoc \
  --proto_path=/app/proto \
  --es_out=/app/web/src/lib/pb \
  --es_opt=target=ts,import_extension=none \
  --connect-es_out=/app/web/src/lib/pb \
  --connect-es_opt=target=ts,import_extension=none \
  --plugin=protoc-gen-connect-es=/app/web/node_modules/.bin/protoc-gen-connect-es \
  --plugin=protoc-gen-es=/app/web/node_modules/.bin/protoc-gen-es \
  $PROTO_FILES
echo "TypeScript proto generation complete!"
