#!/bin/bash

# Thread Art Generator - Infrastructure Setup Script
# This script sets up the initial infrastructure for the staging environment

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="staging"
PROJECT_ID=""
REGION="us-central1"
TERRAFORM_STATE_BUCKET=""
SKIP_BUCKET_CREATION=false

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

# Function to check if required tools are installed
check_dependencies() {
    print_status "Checking dependencies..."
    
    local missing_tools=()
    
    if ! command -v gcloud &> /dev/null; then
        missing_tools+=("gcloud")
    fi
    
    if ! command -v terraform &> /dev/null; then
        missing_tools+=("terraform")
    fi
    
    if [ ${#missing_tools[@]} -ne 0 ]; then
        print_error "Missing required tools: ${missing_tools[*]}"
        echo "Please install the missing tools and try again."
        echo "  - gcloud: https://cloud.google.com/sdk/docs/install"
        echo "  - terraform: https://learn.hashicorp.com/tutorials/terraform/install-cli"
        exit 1
    fi
    
    print_success "All dependencies are installed"
}

# Function to verify GCP authentication
check_gcp_auth() {
    print_status "Checking GCP authentication..."
    
    if ! gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -1 > /dev/null; then
        print_error "No active GCP authentication found"
        echo "Please run: gcloud auth login"
        exit 1
    fi
    
    local active_account=$(gcloud auth list --filter=status:ACTIVE --format="value(account)" | head -1)
    print_success "Authenticated as: $active_account"
}

# Function to set GCP project
set_gcp_project() {
    if [ -z "$PROJECT_ID" ]; then
        print_error "PROJECT_ID is required"
        exit 1
    fi
    
    print_status "Setting GCP project to $PROJECT_ID..."
    
    if ! gcloud config set project "$PROJECT_ID"; then
        print_error "Failed to set project $PROJECT_ID"
        exit 1
    fi
    
    print_success "Project set to $PROJECT_ID"
}

# Function to enable required APIs
enable_apis() {
    print_status "Enabling required GCP APIs..."
    
    local apis=(
        "compute.googleapis.com"
        "cloudresourcemanager.googleapis.com"
        "iam.googleapis.com"
        "serviceusage.googleapis.com"
        "storage.googleapis.com"
        "sql-component.googleapis.com"
        "sqladmin.googleapis.com"
        "run.googleapis.com"
        "artifactregistry.googleapis.com"
        "secretmanager.googleapis.com"
        "redis.googleapis.com"
        "pubsub.googleapis.com"
        "vpcaccess.googleapis.com"
        "servicenetworking.googleapis.com"
        "cloudfunctions.googleapis.com"
        "cloudbuild.googleapis.com"
        "billing.googleapis.com"
    )
    
    for api in "${apis[@]}"; do
        print_status "Enabling $api..."
        if ! gcloud services enable "$api" --quiet; then
            print_warning "Failed to enable $api (might already be enabled)"
        fi
    done
    
    print_success "GCP APIs enabled"
}

# Function to create Terraform state bucket
create_state_bucket() {
    if [ "$SKIP_BUCKET_CREATION" = true ]; then
        print_status "Skipping bucket creation as requested"
        return
    fi
    
    if [ -z "$TERRAFORM_STATE_BUCKET" ]; then
        TERRAFORM_STATE_BUCKET="thread-art-terraform-state-$PROJECT_ID"
    fi
    
    print_status "Creating Terraform state bucket: $TERRAFORM_STATE_BUCKET..."
    
    # Check if bucket already exists
    if gsutil ls -b "gs://$TERRAFORM_STATE_BUCKET" &> /dev/null; then
        print_warning "Bucket $TERRAFORM_STATE_BUCKET already exists"
    else
        # Create bucket
        if ! gsutil mb -p "$PROJECT_ID" -l "$REGION" "gs://$TERRAFORM_STATE_BUCKET"; then
            print_error "Failed to create bucket $TERRAFORM_STATE_BUCKET"
            exit 1
        fi
        
        # Enable versioning
        if ! gsutil versioning set on "gs://$TERRAFORM_STATE_BUCKET"; then
            print_error "Failed to enable versioning on bucket $TERRAFORM_STATE_BUCKET"
            exit 1
        fi
        
        print_success "Created Terraform state bucket: $TERRAFORM_STATE_BUCKET"
    fi
}

# Function to initialize Terraform
init_terraform() {
    print_status "Initializing Terraform..."
    
    local terraform_dir="terraform/environments/$ENVIRONMENT"
    
    if [ ! -d "$terraform_dir" ]; then
        print_error "Terraform directory not found: $terraform_dir"
        exit 1
    fi
    
    cd "$terraform_dir"
    
    # Update backend configuration
    if [ -n "$TERRAFORM_STATE_BUCKET" ]; then
        print_status "Updating backend configuration..."
        cat > backend.tf << EOF
terraform {
  backend "gcs" {
    bucket = "$TERRAFORM_STATE_BUCKET"
    prefix = "$ENVIRONMENT"
  }
}
EOF
    fi
    
    # Initialize Terraform
    if ! terraform init; then
        print_error "Terraform initialization failed"
        exit 1
    fi
    
    print_success "Terraform initialized"
    cd - > /dev/null
}

# Function to create terraform.tfvars if it doesn't exist
create_tfvars() {
    local terraform_dir="terraform/environments/$ENVIRONMENT"
    local tfvars_file="$terraform_dir/terraform.tfvars"
    
    if [ -f "$tfvars_file" ]; then
        print_warning "terraform.tfvars already exists, skipping creation"
        return
    fi
    
    print_status "Creating terraform.tfvars from example..."
    
    if [ ! -f "$terraform_dir/terraform.tfvars.example" ]; then
        print_error "terraform.tfvars.example not found"
        exit 1
    fi
    
    cp "$terraform_dir/terraform.tfvars.example" "$tfvars_file"
    
    # Update with provided values
    sed -i.bak "s/your-gcp-project-id/$PROJECT_ID/g" "$tfvars_file"
    sed -i.bak "s/us-central1/$REGION/g" "$tfvars_file"
    rm "$tfvars_file.bak"
    
    print_warning "Created $tfvars_file - please review and update with your values"
    print_warning "Especially update:"
    echo "  - alert_emails"
    echo "  - github_repository_owner"
    echo "  - firebase_project_id"
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Setup script for Thread Art Generator infrastructure

OPTIONS:
    -p, --project-id PROJECT_ID     GCP Project ID (required)
    -r, --region REGION            GCP Region (default: us-central1)
    -e, --environment ENV          Environment name (default: staging)
    -b, --bucket BUCKET           Terraform state bucket name (default: auto-generated)
    -s, --skip-bucket             Skip bucket creation
    -h, --help                    Show this help message

EXAMPLES:
    $0 --project-id my-project-123
    $0 --project-id my-project-123 --region us-west1
    $0 --project-id my-project-123 --skip-bucket

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -p|--project-id)
            PROJECT_ID="$2"
            shift 2
            ;;
        -r|--region)
            REGION="$2"
            shift 2
            ;;
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -b|--bucket)
            TERRAFORM_STATE_BUCKET="$2"
            shift 2
            ;;
        -s|--skip-bucket)
            SKIP_BUCKET_CREATION=true
            shift
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
    echo "Thread Art Generator - Infrastructure Setup"
    echo "=========================================="
    echo
    
    # Validate required parameters
    if [ -z "$PROJECT_ID" ]; then
        print_error "Project ID is required"
        usage
        exit 1
    fi
    
    print_status "Configuration:"
    echo "  Project ID: $PROJECT_ID"
    echo "  Region: $REGION"
    echo "  Environment: $ENVIRONMENT"
    echo "  State Bucket: ${TERRAFORM_STATE_BUCKET:-auto-generated}"
    echo
    
    # Run setup steps
    check_dependencies
    check_gcp_auth
    set_gcp_project
    enable_apis
    create_state_bucket
    init_terraform
    create_tfvars
    
    echo
    print_success "Infrastructure setup completed!"
    echo
    print_status "Next steps:"
    echo "1. Review and update terraform/environments/$ENVIRONMENT/terraform.tfvars"
    echo "2. Run: cd terraform/environments/$ENVIRONMENT"
    echo "3. Run: terraform plan"
    echo "4. Run: terraform apply"
    echo
    print_warning "Remember: This will deploy infrastructure that may incur costs!"
    print_warning "The budget is set to \$50/month with auto-shutdown enabled."
}

# Run main function
main "$@"