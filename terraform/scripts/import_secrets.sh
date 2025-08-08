#!/usr/bin/env bash

# Script to import existing secrets into Terraform state
# This handles cases where secrets already exist in GCP
# 
# This script should be run from the terraform/environments/staging directory

set -e

# Configuration
PROJECT_ID="538844318206"
ENVIRONMENT="staging"

echo "Thread Art Generator - Secret Import Script"
echo "==========================================="
echo ""
echo "Importing existing secrets for:"
echo "  Project: $PROJECT_ID"
echo "  Environment: $ENVIRONMENT"
echo ""

# Define the secrets to import
declare -A SECRETS=(
    ["token-symmetric-key-staging"]="module.secrets.google_secret_manager_secret.token_symmetric_key"
    ["internal-api-key-staging"]="module.secrets.google_secret_manager_secret.internal_api_key"
    ["cookie-hash-key-staging"]="module.secrets.google_secret_manager_secret.cookie_hash_key"
    ["cookie-block-key-staging"]="module.secrets.google_secret_manager_secret.cookie_block_key"
)

# Check if we're in the right directory
if [[ ! -f "terraform.tfvars" ]] || [[ ! -f "main.tf" ]]; then
    echo "❌ Error: This script must be run from the terraform/environments/staging directory"
    echo "   Current directory: $(pwd)"
    echo ""
    echo "Please run:"
    echo "   cd terraform/environments/staging"
    echo "   ../../../terraform/scripts/import_secrets.sh"
    exit 1
fi

echo "✓ Verified correct directory: $(pwd)"
echo ""

# Initialize Terraform if needed
if [[ ! -d ".terraform" ]]; then
    echo "🔧 Initializing Terraform..."
    terraform init
    echo ""
fi

# Import each secret
echo "🔐 Importing secrets..."
echo ""

for secret_name in "${!SECRETS[@]}"; do
    terraform_resource="${SECRETS[$secret_name]}"
    secret_path="projects/${PROJECT_ID}/secrets/${secret_name}"
    
    echo "Importing: $secret_name"
    echo "  Resource: $terraform_resource"
    echo "  Path: $secret_path"
    
    # Check if the resource already exists in state
    if terraform state show "$terraform_resource" >/dev/null 2>&1; then
        echo "  ✓ Already exists in state, skipping"
    else
        echo "  → Importing..."
        if terraform import "$terraform_resource" "$secret_path" >/dev/null 2>&1; then
            echo "  ✅ Successfully imported"
        else
            echo "  ❌ Failed to import"
            echo "     This might be because:"
            echo "     1. The secret doesn't exist in GCP"
            echo "     2. Authentication issues"
            echo "     3. Incorrect project ID"
            echo ""
            echo "     You can create missing secrets using:"
            echo "     ./terraform/scripts/secrets.sh --project-id $PROJECT_ID --action init"
            exit 1
        fi
    fi
    echo ""
done

echo "🎉 All secrets imported successfully!"
echo ""
echo "Next steps:"
echo "1. Run 'terraform plan' to verify the import worked correctly"
echo "2. Run 'terraform apply' to proceed with your deployment"
echo ""
echo "If you need to set secret values, use:"
echo "  ./terraform/scripts/secrets.sh --project-id $PROJECT_ID --action generate"