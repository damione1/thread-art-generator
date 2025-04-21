#!/bin/bash
set -e

echo "Starting MinIO setup..."

# Retry logic for connecting to MinIO
MAX_RETRIES=10
RETRY_INTERVAL=3
SUCCESS=false

for i in $(seq 1 $MAX_RETRIES); do
  echo "Attempt $i: Connecting to MinIO..."

  if curl -s --connect-timeout 5 http://localhost:9000/minio/health/live > /dev/null; then
    echo "✅ MinIO is running"
    SUCCESS=true
    break
  else
    echo "⏳ MinIO not ready, waiting $RETRY_INTERVAL seconds..."
    sleep $RETRY_INTERVAL
  fi
done

if ! $SUCCESS; then
  echo "❌ Failed to connect to MinIO after $MAX_RETRIES attempts"
  exit 1
fi

# Get environment variables
if [ -f .env ]; then
  export $(grep -v '^#' .env | grep -e STORAGE_ | xargs)
else
  echo "❌ .env file not found, cannot configure MinIO"
  exit 1
fi

# Install mc if not available
MC_PATH=$(which mc || echo "")
if [ -z "$MC_PATH" ]; then
  echo "Installing MinIO client (mc)..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    brew install minio/stable/mc
  else
    echo "❌ Please install MinIO client (mc) manually"
    exit 1
  fi
fi

# Configure MinIO client
echo "Configuring MinIO client..."
mc alias set local http://localhost:9000 "$STORAGE_ACCESS_KEY" "$STORAGE_SECRET_KEY" > /dev/null

# Create bucket if it doesn't exist
echo "Checking for bucket: $STORAGE_BUCKET"
if ! mc ls local 2>/dev/null | grep -q "$STORAGE_BUCKET"; then
  echo "Creating bucket: $STORAGE_BUCKET"
  mc mb local/$STORAGE_BUCKET > /dev/null
fi

# Set anonymous policy for bucket
echo "Setting download policy for bucket: $STORAGE_BUCKET"
mc anonymous set download local/$STORAGE_BUCKET > /dev/null

echo "✅ MinIO setup complete!"
