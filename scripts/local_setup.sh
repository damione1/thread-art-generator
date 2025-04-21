#!/bin/bash
set -e

# Colors for output
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
NC='\033[0m' # No Color

# Directory of this script
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
PROJECT_ROOT="$(dirname "$SCRIPT_DIR")"

echo -e "${GREEN}=== Thread Art Generator - Local Development Setup ===${NC}"

# Function to check if a command exists
function command_exists() {
    command -v "$1" &> /dev/null
}

# Function to check and setup required tools
function setup_tools() {
    echo -e "\n${YELLOW}Checking required tools...${NC}"

    # Check for Docker
    if ! command_exists docker; then
        echo -e "${RED}Docker is not installed. Please install Docker Desktop or Docker Engine.${NC}"
        exit 1
    fi
    echo -e "✅ Docker is installed"

    # Check for docker-compose
    if ! command_exists docker-compose; then
        echo -e "${RED}docker-compose is not installed. Please install it.${NC}"
        exit 1
    fi
    echo -e "✅ docker-compose is installed"

    # Check for Tilt
    if ! command_exists tilt; then
        echo -e "${YELLOW}Tilt is not installed. Would you like to install it? (y/n)${NC}"
        read -r install_tilt
        if [[ "$install_tilt" =~ ^[Yy]$ ]]; then
            if [[ "$OSTYPE" == "darwin"* ]]; then
                brew install tilt
            else
                echo -e "${RED}Please install Tilt manually: https://docs.tilt.dev/install.html${NC}"
                exit 1
            fi
        else
            echo -e "${RED}Tilt is required for local development.${NC}"
            exit 1
        fi
    fi
    echo -e "✅ Tilt is installed"

    # Check for mkcert
    if ! command_exists mkcert; then
        echo -e "${YELLOW}mkcert is not installed. Installing...${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install mkcert
        else
            echo -e "${RED}Unsupported OS, please install mkcert manually${NC}"
            exit 1
        fi
    fi
    echo -e "✅ mkcert is installed"
}

# Function to setup SSL certificates
function setup_ssl() {
    echo -e "\n${YELLOW}Setting up SSL certificates...${NC}"

    # Install the local CA
    mkcert -install

    # Create certs directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/certs"

    # Generate certificates for tag.local and storage.tag.local
    echo "Generating certificates for tag.local and storage.tag.local..."
    mkcert -cert-file "$PROJECT_ROOT/certs/tag.local.crt" -key-file "$PROJECT_ROOT/certs/tag.local.key" tag.local "*.tag.local" storage.tag.local

    # Add domains to /etc/hosts if not already present
    if ! grep -q "tag.local" /etc/hosts; then
        echo "Adding domains to /etc/hosts..."
        echo "You might be prompted for your password to modify /etc/hosts"
        sudo sh -c "echo '127.0.0.1 tag.local' >> /etc/hosts"
        sudo sh -c "echo '127.0.0.1 storage.tag.local' >> /etc/hosts"
    fi

    echo -e "✅ SSL certificates setup complete"
}

# Setup environment variables
function setup_env() {
    echo -e "\n${YELLOW}Setting up environment variables...${NC}"

    if [ ! -f "$PROJECT_ROOT/.env" ]; then
        echo "Creating .env file from template..."
        cp "$PROJECT_ROOT/.env.sample" "$PROJECT_ROOT/.env"

        # Generate a symmetric key for tokens
        TOKEN_SYMMETRIC_KEY=$(openssl rand -hex 16)
        echo "TOKEN_SYMMETRIC_KEY=$TOKEN_SYMMETRIC_KEY" >> "$PROJECT_ROOT/.env"

        # Generate a secret for NextAuth
        NEXTAUTH_SECRET=$(openssl rand -hex 32)
        echo "NEXTAUTH_SECRET=$NEXTAUTH_SECRET" >> "$PROJECT_ROOT/.env"

        echo -e "✅ .env file created with generated keys"
    else
        echo -e "✅ .env file already exists, skipping"
    fi
}

# Build the CLI
function build_cli() {
    echo -e "\n${YELLOW}Building CLI tool...${NC}"
    mkdir -p "$PROJECT_ROOT/build"
    (cd "$PROJECT_ROOT" && go build -o build/cli cmd/cli/main.go)
    chmod +x "$PROJECT_ROOT/build/cli"
    echo -e "✅ CLI built successfully at build/cli"
}

# Main setup logic
setup_tools
setup_ssl
setup_env
build_cli

echo -e "\n${GREEN}Setup complete! You can now start the development environment with:${NC}"
echo -e "${YELLOW}tilt up${NC}"
echo -e "\n${GREEN}To reset the environment:${NC}"
echo -e "${YELLOW}tilt down && tilt up${NC}"
echo -e "\n${GREEN}Happy coding!${NC}"
