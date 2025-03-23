#!/bin/bash
set -e

# Directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

# Load environment variables from .env file
if [ -f "$PROJECT_ROOT/.env" ]; then
  echo "Loading environment variables from .env file"
  export $(grep -v '^#' "$PROJECT_ROOT/.env" | xargs)
else
  echo "Warning: .env file not found in $PROJECT_ROOT"
fi

# Build the CLI if it doesn't exist
if [ ! -f "$PROJECT_ROOT/build/cli" ] || [ "$1" == "build" ]; then
  echo "Building Thread Art CLI..."
  mkdir -p "$PROJECT_ROOT/build"
  (cd "$PROJECT_ROOT" && go build -o build/cli cmd/cli/main.go)
  chmod +x "$PROJECT_ROOT/build/cli"
  echo "CLI built successfully at build/cli"

  # Exit if only building
  if [ "$1" == "build" ]; then
    exit 0
  fi
fi

# Check if AUTH0 environment variables are set
if [ -z "$AUTH0_DOMAIN" ] || [ -z "$AUTH0_CLIENT_ID" ] || [ -z "$AUTH0_CLIENT_SECRET" ] || [ -z "$AUTH0_AUDIENCE" ]; then
  echo "Error: Missing required AUTH0 environment variables."
  echo "Make sure your .env file contains:"
  echo "  AUTH0_DOMAIN"
  echo "  AUTH0_CLIENT_ID"
  echo "  AUTH0_CLIENT_SECRET"
  echo "  AUTH0_AUDIENCE"
  exit 1
fi

# Run the CLI with arguments
"$PROJECT_ROOT/build/cli" "$@"
