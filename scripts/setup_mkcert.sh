#!/bin/bash
set -e

# Check if mkcert is installed
if ! command -v mkcert &> /dev/null; then
  echo "mkcert is not installed, installing..."
  if [[ "$OSTYPE" == "darwin"* ]]; then
    brew install mkcert
  else
    echo "Unsupported OS, please install mkcert manually"
    exit 1
  fi
fi

# Install the local CA
mkcert -install

# Create certs directory if it doesn't exist
mkdir -p ./certs

# Generate certificates for tag.local
echo "Generating certificates for tag.local..."
mkcert -cert-file ./certs/tag.local.crt -key-file ./certs/tag.local.key tag.local "*.tag.local" storage.tag.local

# Add domains to /etc/hosts if not already present
if ! grep -q "tag.local" /etc/hosts; then
  echo "Adding domains to /etc/hosts..."
  echo "You might be prompted for your password to modify /etc/hosts"
  sudo sh -c 'echo "127.0.0.1 tag.local" >> /etc/hosts'
  sudo sh -c 'echo "127.0.0.1 storage.tag.local" >> /etc/hosts'
fi

echo "Setup complete! Your certificates are in the ./certs directory"
