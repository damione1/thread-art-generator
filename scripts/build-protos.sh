#!/bin/bash
set -e

# Find all .proto files in the /app/proto directory and its subdirectories
PROTO_FILES=$(find /app/proto -name "*.proto")

echo "Generating Go proto files..."
rm -rf /app/core/pb
mkdir -p /app/core/pb

# Build proto paths dynamically based on what exists
PROTO_PATHS="--proto_path=/app/proto"

# Add additional paths only if they exist
if [ -d "/app/proto/third_party" ]; then
  PROTO_PATHS="$PROTO_PATHS --proto_path=/app/proto/third_party"
fi

if [ -d "/app/proto/google" ]; then
  PROTO_PATHS="$PROTO_PATHS --proto_path=/app/proto/google"
fi

protoc \
  $PROTO_PATHS \
  --go_out=paths=source_relative:/app/core/pb \
  --go-grpc_out=paths=source_relative:/app/core/pb \
  --grpc-gateway_out=paths=source_relative:/app/core/pb \
  $PROTO_FILES

# Format generated Go files with consistent import ordering
echo "Formatting Go files with consistent import ordering..."
if command -v goimports &> /dev/null; then
  find /app/core/pb -name "*.go" -exec goimports -w -local github.com/Damione1/thread-art-generator {} \;
else
  echo "Warning: goimports not found. Import ordering may be inconsistent."
fi
echo "Go proto generation complete!"

echo "Generating TypeScript proto files..."
rm -rf /app/web/src/lib/pb/*
mkdir -p /app/web/src/lib/pb

protoc \
  $PROTO_PATHS \
  --es_out=/app/web/src/lib/pb \
  --es_opt=target=ts,import_extension=none \
  --connect-es_out=/app/web/src/lib/pb \
  --connect-es_opt=target=ts,import_extension=none \
  --plugin=protoc-gen-connect-es=/app/web/node_modules/.bin/protoc-gen-connect-es \
  --plugin=protoc-gen-es=/app/web/node_modules/.bin/protoc-gen-es \
  $PROTO_FILES
echo "TypeScript proto generation complete!"
