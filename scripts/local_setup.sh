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

    # Check for Node.js and npm (needed for Tailwind CSS)
    if ! command_exists npm; then
        echo -e "${YELLOW}npm is not installed. Installing for Tailwind CSS...${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install node
        else
            echo -e "${RED}Please install Node.js and npm manually: https://nodejs.org/${NC}"
            exit 1
        fi
    fi
    echo -e "✅ npm is installed"

    # Check for buf CLI
    if ! command_exists buf; then
        echo -e "${YELLOW}buf CLI is not installed. Installing...${NC}"
        if [[ "$OSTYPE" == "darwin"* ]]; then
            brew install bufbuild/buf/buf
        else
            echo -e "${RED}Please install buf CLI manually: https://buf.build/docs/installation${NC}"
            exit 1
        fi
    fi
    echo -e "✅ buf CLI is installed"

    # Check for Go
    if ! command_exists go; then
        echo -e "${RED}Go is not installed. Please install Go 1.22+: https://golang.org/dl/${NC}"
        exit 1
    fi
    echo -e "✅ Go is installed"

    # Check Go version
    GO_VERSION=$(go version | grep -o 'go[0-9]\+\.[0-9]\+' | sed 's/go//')
    if [ "$(printf '%s\n' "1.22" "$GO_VERSION" | sort -V | head -n1)" != "1.22" ]; then
        echo -e "${RED}Go version 1.22+ is required. Current version: $GO_VERSION${NC}"
        exit 1
    fi
    echo -e "✅ Go version $GO_VERSION is compatible"
}

# Function to setup SSL certificates
function setup_ssl() {
    echo -e "\n${YELLOW}Setting up SSL certificates...${NC}"

    # Install the local CA
    mkcert -install

    # Create certs directory if it doesn't exist
    mkdir -p "$PROJECT_ROOT/certs"

    # Generate certificates for tag.local, front.tag.local and storage.tag.local
    echo "Generating certificates for tag.local, front.tag.local and storage.tag.local..."
    mkcert -cert-file "$PROJECT_ROOT/certs/tag.local.crt" -key-file "$PROJECT_ROOT/certs/tag.local.key" tag.local "*.tag.local" front.tag.local storage.tag.local

    # Add domains to /etc/hosts if not already present
    if ! grep -q "tag.local" /etc/hosts; then
        echo "Adding domains to /etc/hosts..."
        echo "You might be prompted for your password to modify /etc/hosts"
        sudo sh -c "echo '127.0.0.1 tag.local' >> /etc/hosts"
        sudo sh -c "echo '127.0.0.1 front.tag.local' >> /etc/hosts"
        sudo sh -c "echo '127.0.0.1 storage.tag.local' >> /etc/hosts"
    fi

    # Check if front.tag.local is in hosts, add it if not
    if ! grep -q "front.tag.local" /etc/hosts; then
        echo "Adding front.tag.local to /etc/hosts..."
        echo "You might be prompted for your password to modify /etc/hosts"
        sudo sh -c "echo '127.0.0.1 front.tag.local' >> /etc/hosts"
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

# Setup client frontend
function setup_frontend() {
    echo -e "\n${YELLOW}Setting up Go+HTMX frontend...${NC}"

    # Get Go path
    GOPATH=$(go env GOPATH)
    GOBIN=$GOPATH/bin

    # Install templ
    echo -e "Installing templ..."
    go install github.com/a-h/templ/cmd/templ@latest
    if ! command_exists $GOBIN/templ; then
        echo -e "${RED}Failed to install templ${NC}"
        exit 1
    fi
    echo -e "✅ templ installed successfully"

    # Install Tailwind CSS dependencies
    echo -e "Installing Tailwind CSS dependencies..."
    (cd "$PROJECT_ROOT/client" && npm install)
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to install Tailwind CSS dependencies${NC}"
        exit 1
    fi
    echo -e "✅ Tailwind CSS dependencies installed successfully"

    # Build Tailwind CSS
    echo -e "Building Tailwind CSS..."
    (cd "$PROJECT_ROOT/client" && mkdir -p ./public/css && npx tailwindcss -i ./styles/input.css -o ./public/css/tailwind.css --minify)
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to build Tailwind CSS${NC}"
        exit 1
    fi
    echo -e "✅ Tailwind CSS built successfully"

    # Generate templ files
    echo -e "Generating templ templates..."
    (cd "$PROJECT_ROOT/client" && $GOBIN/templ generate ./internal/templates)
    if [ $? -ne 0 ]; then
        echo -e "${RED}Failed to generate templ templates${NC}"
        exit 1
    fi
    echo -e "✅ Templates generated successfully"

    echo -e "✅ Go+HTMX frontend setup complete"
}

# Setup protocol buffer tools
function setup_proto_tools() {
    echo -e "\n${YELLOW}Setting up protocol buffer tools...${NC}"

    # Get Go path
    GOPATH=$(go env GOPATH)
    GOBIN=$GOPATH/bin

    # Install protoc-gen-go
    echo -e "Installing protoc-gen-go..."
    go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
    if [ ! -f "$GOBIN/protoc-gen-go" ]; then
        echo -e "${RED}Failed to install protoc-gen-go${NC}"
        exit 1
    fi
    echo -e "✅ protoc-gen-go installed successfully"

    # Install protoc-gen-connect-go
    echo -e "Installing protoc-gen-connect-go..."
    go install connectrpc.com/connect/cmd/protoc-gen-connect-go@latest
    if [ ! -f "$GOBIN/protoc-gen-connect-go" ]; then
        echo -e "${RED}Failed to install protoc-gen-connect-go${NC}"
        exit 1
    fi
    echo -e "✅ protoc-gen-connect-go installed successfully"

    # Install protoc-gen-openapiv2
    echo -e "Installing protoc-gen-openapiv2..."
    go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@latest
    if [ ! -f "$GOBIN/protoc-gen-openapiv2" ]; then
        echo -e "${RED}Failed to install protoc-gen-openapiv2${NC}"
        exit 1
    fi
    echo -e "✅ protoc-gen-openapiv2 installed successfully"

    echo -e "✅ Protocol buffer tools setup complete"
    echo -e "${YELLOW}Note: Protocol buffer tools are installed in $GOBIN${NC}"
    echo -e "${YELLOW}Make sure $GOBIN is in your PATH for direct CLI access${NC}"
}

# Main setup logic
setup_tools
setup_ssl
setup_env
setup_proto_tools
build_cli
setup_frontend

echo -e "\n${GREEN}Setup complete! You can now start the development environment with:${NC}"
echo -e "${YELLOW}tilt up${NC}"
echo -e "\n${GREEN}To reset the environment:${NC}"
echo -e "${YELLOW}tilt down && tilt up${NC}"
echo -e "\n${GREEN}To regenerate protocol buffers (if needed):${NC}"
echo -e "${YELLOW}make proto${NC}"
echo -e "\n${GREEN}Happy coding!${NC}"
