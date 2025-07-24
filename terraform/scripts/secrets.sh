#!/bin/bash

# Thread Art Generator - Secret Management Script
# This script helps manage secrets in Google Secret Manager

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
PROJECT_ID=""
ENVIRONMENT="staging"
ACTION="list"
SECRET_NAME=""
SECRET_VALUE=""
SECRET_FILE=""

# Function to print colored output
print_status() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

print_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

print_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Function to check if gcloud is authenticated
check_gcp_auth() {
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -1 > /dev/null; then
        print_error "No active GCP authentication found"
        echo "Please run: gcloud auth login"
        exit 1
    fi
}

# Function to set GCP project
set_gcp_project() {
    if [ -z "$PROJECT_ID" ]; then
        print_error "PROJECT_ID is required"
        exit 1
    fi
    
    gcloud config set project "$PROJECT_ID" > /dev/null
    print_status "Using project: $PROJECT_ID"
}

# Function to list all secrets
list_secrets() {
    print_status "Listing secrets for environment: $ENVIRONMENT"
    
    gcloud secrets list \
        --filter="labels.environment=$ENVIRONMENT" \
        --format="table(name.basename(),labels.purpose:sort=1,createTime)" \
        --sort-by="labels.purpose"
}

# Function to create a secret
create_secret() {
    if [ -z "$SECRET_NAME" ]; then
        print_error "Secret name is required for create action"
        exit 1
    fi
    
    local full_secret_name="${SECRET_NAME}-${ENVIRONMENT}"
    
    print_status "Creating secret: $full_secret_name"
    
    # Check if secret already exists
    if gcloud secrets describe "$full_secret_name" &> /dev/null; then
        print_warning "Secret $full_secret_name already exists"
        return 0
    fi
    
    # Create the secret
    if ! gcloud secrets create "$full_secret_name" \
        --labels="environment=$ENVIRONMENT,purpose=${SECRET_NAME//-/_}" \
        --replication-policy="automatic"; then
        print_error "Failed to create secret $full_secret_name"
        exit 1
    fi
    
    print_success "Created secret: $full_secret_name"
}

# Function to set secret value
set_secret() {
    if [ -z "$SECRET_NAME" ]; then
        print_error "Secret name is required for set action"
        exit 1
    fi
    
    local full_secret_name="${SECRET_NAME}-${ENVIRONMENT}"
    local value_source=""
    
    # Determine value source
    if [ -n "$SECRET_VALUE" ]; then
        value_source="$SECRET_VALUE"
    elif [ -n "$SECRET_FILE" ]; then
        if [ ! -f "$SECRET_FILE" ]; then
            print_error "Secret file not found: $SECRET_FILE"
            exit 1
        fi
        value_source=$(cat "$SECRET_FILE")
    else
        print_status "Enter secret value (input will be hidden):"
        read -s value_source
        echo
    fi
    
    if [ -z "$value_source" ]; then
        print_error "Secret value cannot be empty"
        exit 1
    fi
    
    print_status "Setting value for secret: $full_secret_name"
    
    # Create secret if it doesn't exist
    if ! gcloud secrets describe "$full_secret_name" &> /dev/null; then
        create_secret
    fi
    
    # Set the secret value
    if ! echo "$value_source" | gcloud secrets versions add "$full_secret_name" --data-file=-; then
        print_error "Failed to set secret value for $full_secret_name"
        exit 1
    fi
    
    print_success "Set value for secret: $full_secret_name"
}

# Function to get secret value
get_secret() {
    if [ -z "$SECRET_NAME" ]; then
        print_error "Secret name is required for get action"
        exit 1
    fi
    
    local full_secret_name="${SECRET_NAME}-${ENVIRONMENT}"
    
    print_status "Getting value for secret: $full_secret_name"
    
    if ! gcloud secrets versions access latest --secret="$full_secret_name"; then
        print_error "Failed to get secret value for $full_secret_name"
        exit 1
    fi
}

# Function to delete a secret
delete_secret() {
    if [ -z "$SECRET_NAME" ]; then
        print_error "Secret name is required for delete action"
        exit 1
    fi
    
    local full_secret_name="${SECRET_NAME}-${ENVIRONMENT}"
    
    print_warning "This will permanently delete secret: $full_secret_name"
    read -p "Are you sure? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Operation cancelled"
        return 0
    fi
    
    print_status "Deleting secret: $full_secret_name"
    
    if ! gcloud secrets delete "$full_secret_name" --quiet; then
        print_error "Failed to delete secret $full_secret_name"
        exit 1
    fi
    
    print_success "Deleted secret: $full_secret_name"
}

# Function to initialize all required secrets
init_secrets() {
    print_status "Initializing all required secrets for environment: $ENVIRONMENT"
    
    local secrets=(
        "token-symmetric-key"
        "internal-api-key"
        "cookie-hash-key"
        "cookie-block-key"
    )
    
    for secret in "${secrets[@]}"; do
        SECRET_NAME="$secret"
        create_secret
    done
    
    print_success "All secrets initialized"
    print_warning "Remember to set values for each secret using:"
    print_warning "  $0 --action set --secret <secret-name>"
}

# Function to generate secure values for secrets
generate_values() {
    print_status "Generating secure values for secrets..."
    
    local temp_dir=$(mktemp -d)
    
    # Generate values
    local token_key=$(openssl rand -base64 64)
    local api_key=$(openssl rand -base64 32)
    local hash_key=$(openssl rand -base64 32)
    local block_key=$(openssl rand -base64 24)
    
    # Create temporary files
    echo "$token_key" > "$temp_dir/token-symmetric-key"
    echo "$api_key" > "$temp_dir/internal-api-key"
    echo "$hash_key" > "$temp_dir/cookie-hash-key"
    echo "$block_key" > "$temp_dir/cookie-block-key"
    
    # Set each secret
    for secret_file in "$temp_dir"/*; do
        local secret_name=$(basename "$secret_file")
        SECRET_NAME="$secret_name"
        SECRET_FILE="$secret_file"
        set_secret
        SECRET_FILE=""  # Reset for next iteration
    done
    
    # Cleanup
    rm -rf "$temp_dir"
    
    print_success "Generated and set all secret values"
}

# Function to export secrets to env file
export_secrets() {
    print_status "Exporting secrets to .env file..."
    
    local env_file=".env.secrets"
    
    # Warning about sensitive data
    print_warning "This will create $env_file with sensitive data"
    read -p "Continue? (y/N): " -n 1 -r
    echo
    
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        print_status "Operation cancelled"
        return 0
    fi
    
    # Create env file
    cat > "$env_file" << EOF
# Generated secrets for $ENVIRONMENT environment
# WARNING: This file contains sensitive data - do not commit to git
# Note: Database authentication uses IAM - no passwords needed

TOKEN_SYMMETRIC_KEY=$(SECRET_NAME="token-symmetric-key" get_secret 2>/dev/null || echo "")
INTERNAL_API_KEY=$(SECRET_NAME="internal-api-key" get_secret 2>/dev/null || echo "")
COOKIE_HASH_KEY=$(SECRET_NAME="cookie-hash-key" get_secret 2>/dev/null || echo "")
COOKIE_BLOCK_KEY=$(SECRET_NAME="cookie-block-key" get_secret 2>/dev/null || echo "")
EOF
    
    chmod 600 "$env_file"
    print_success "Secrets exported to $env_file"
    print_warning "Remember to add $env_file to .gitignore"
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Secret management script for Thread Art Generator

OPTIONS:
    -p, --project-id PROJECT_ID    GCP Project ID (required)
    -e, --environment ENV          Environment name (default: staging)
    -a, --action ACTION           Action to perform (default: list)
    -s, --secret SECRET_NAME      Secret name (for create/set/get/delete actions)
    -v, --value SECRET_VALUE      Secret value (for set action)
    -f, --file SECRET_FILE        File containing secret value (for set action)
    -h, --help                    Show this help message

ACTIONS:
    list         List all secrets for environment
    create       Create a new secret
    set          Set secret value (interactive, from value, or from file)
    get          Get secret value
    delete       Delete a secret
    init         Initialize all required secrets
    generate     Generate secure values for all secrets
    export       Export secrets to .env file

EXAMPLES:
    $0 --project-id my-project --action list
    $0 --project-id my-project --action create --secret token-symmetric-key
    $0 --project-id my-project --action set --secret token-symmetric-key
    $0 --project-id my-project --action set --secret token-symmetric-key --value mytoken
    $0 --project-id my-project --action set --secret token-symmetric-key --file token.txt
    $0 --project-id my-project --action get --secret token-symmetric-key
    $0 --project-id my-project --action init
    $0 --project-id my-project --action generate

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--project-id)
            PROJECT_ID="$2"
            shift 2
            ;;
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -a|--action)
            ACTION="$2"
            shift 2
            ;;
        -s|--secret)
            SECRET_NAME="$2"
            shift 2
            ;;
        -v|--value)
            SECRET_VALUE="$2"
            shift 2
            ;;
        -f|--file)
            SECRET_FILE="$2"
            shift 2
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        *)
            print_error "Unknown option: $1"
            usage
            exit 1
            ;;
    esac
done

# Main execution
main() {
    echo "Thread Art Generator - Secret Management"
    echo "======================================="
    echo
    
    # Validate required parameters
    if [ -z "$PROJECT_ID" ]; then
        print_error "Project ID is required"
        usage
        exit 1
    fi
    
    print_status "Configuration:"
    echo "  Project ID: $PROJECT_ID"
    echo "  Environment: $ENVIRONMENT"
    echo "  Action: $ACTION"
    echo "  Secret: ${SECRET_NAME:-N/A}"
    echo
    
    # Check authentication and set project
    check_gcp_auth
    set_gcp_project
    
    # Execute requested action
    case $ACTION in
        list)
            list_secrets
            ;;
        create)
            create_secret
            ;;
        set)
            set_secret
            ;;
        get)
            get_secret
            ;;
        delete)
            delete_secret
            ;;
        init)
            init_secrets
            ;;
        generate)
            generate_values
            ;;
        export)
            export_secrets
            ;;
        *)
            print_error "Unknown action: $ACTION"
            usage
            exit 1
            ;;
    esac
}

# Run main function
main "$@"