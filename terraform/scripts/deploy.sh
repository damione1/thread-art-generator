#!/bin/bash

# Thread Art Generator - Deployment Script
# This script handles Terraform deployment with safety checks

set -euo pipefail

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Default values
ENVIRONMENT="staging"
ACTION="plan"
AUTO_APPROVE=false
DESTROY=false
TARGET=""
VAR_FILE=""

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

# Function to check if Terraform is initialized
check_terraform_init() {
    if [ ! -d ".terraform" ]; then
        print_error "Terraform not initialized. Run terraform init first."
        exit 1
    fi
}

# Function to validate Terraform configuration
validate_terraform() {
    print_status "Validating Terraform configuration..."
    
    if ! terraform validate; then
        print_error "Terraform configuration validation failed"
        exit 1
    fi
    
    print_success "Terraform configuration is valid"
}

# Function to run terraform plan
terraform_plan() {
    print_status "Running Terraform plan..."
    
    local plan_args=()
    
    if [ -n "$VAR_FILE" ]; then
        plan_args+=("-var-file=$VAR_FILE")
    fi
    
    if [ -n "$TARGET" ]; then
        plan_args+=("-target=$TARGET")
    fi
    
    if [ "$DESTROY" = true ]; then
        plan_args+=("-destroy")
    fi
    
    # Save plan to file
    local plan_file="terraform-$(date +%Y%m%d-%H%M%S).tfplan"
    plan_args+=("-out=$plan_file")
    
    if ! terraform plan "${plan_args[@]}"; then
        print_error "Terraform plan failed"
        exit 1
    fi
    
    print_success "Terraform plan completed successfully"
    echo "Plan saved to: $plan_file"
    
    # Show cost estimation if available
    if command -v infracost &> /dev/null; then
        print_status "Generating cost estimate..."
        infracost breakdown --path . || print_warning "Cost estimation failed"
    fi
    
    return 0
}

# Function to run terraform apply
terraform_apply() {
    print_status "Running Terraform apply..."
    
    local apply_args=()
    
    if [ "$AUTO_APPROVE" = true ]; then
        apply_args+=("-auto-approve")
    fi
    
    if [ -n "$VAR_FILE" ]; then
        apply_args+=("-var-file=$VAR_FILE")
    fi
    
    if [ -n "$TARGET" ]; then
        apply_args+=("-target=$TARGET")
    fi
    
    if [ "$DESTROY" = true ]; then
        apply_args+=("-destroy")
        if [ "$AUTO_APPROVE" != true ]; then
            print_warning "This will DESTROY infrastructure. Type 'yes' to confirm."
        fi
    fi
    
    if ! terraform apply "${apply_args[@]}"; then
        print_error "Terraform apply failed"
        exit 1
    fi
    
    print_success "Terraform apply completed successfully"
}

# Function to run terraform output
show_outputs() {
    print_status "Terraform outputs:"
    terraform output -json | jq -r '
        to_entries[] |
        if .value.sensitive then
            "\(.key): <sensitive>"
        else
            "\(.key): \(.value.value)"
        end
    ' 2>/dev/null || terraform output
}

# Function to check for drift
check_drift() {
    print_status "Checking for configuration drift..."
    
    if terraform plan -detailed-exitcode -refresh-only > /dev/null; then
        print_success "No drift detected"
    else
        case $? in
            1)
                print_error "Error checking for drift"
                exit 1
                ;;
            2)
                print_warning "Drift detected! Infrastructure differs from configuration"
                print_status "Run 'terraform plan -refresh-only' to see details"
                ;;
        esac
    fi
}

# Function to backup state
backup_state() {
    print_status "Backing up Terraform state..."
    
    local backup_dir="backups/$(date +%Y%m%d-%H%M%S)"
    mkdir -p "$backup_dir"
    
    if terraform state pull > "$backup_dir/terraform.tfstate.backup"; then
        print_success "State backed up to: $backup_dir/terraform.tfstate.backup"
    else
        print_warning "Failed to backup state"
    fi
}

# Function to show resource count and estimated costs
show_summary() {
    print_status "Infrastructure Summary:"
    
    # Count resources by type
    terraform state list | cut -d. -f1 | sort | uniq -c | sort -nr | head -10
    
    echo
    print_status "Recent deployments:"
    ls -la terraform-*.tfplan 2>/dev/null | tail -5 || echo "No recent plan files found"
}

# Function to show usage
usage() {
    cat << EOF
Usage: $0 [OPTIONS]

Deployment script for Thread Art Generator infrastructure

OPTIONS:
    -e, --environment ENV      Environment name (default: staging)
    -a, --action ACTION       Action to perform: plan, apply, output, check, summary (default: plan)
    -f, --var-file FILE       Terraform variables file
    -t, --target RESOURCE     Target specific resource
    -y, --auto-approve        Auto-approve apply (use with caution)
    -d, --destroy             Destroy infrastructure
    -h, --help                Show this help message

ACTIONS:
    plan          Run terraform plan
    apply         Run terraform apply
    output        Show terraform outputs
    check         Check for configuration drift
    summary       Show infrastructure summary
    backup        Backup current state

EXAMPLES:
    $0                                    # Run plan for staging
    $0 --action apply                     # Apply changes to staging
    $0 --action apply --auto-approve      # Apply without confirmation (CI/CD)
    $0 --action destroy --auto-approve    # Destroy infrastructure
    $0 --action check                     # Check for drift
    $0 --target module.database           # Target specific module

SAFETY FEATURES:
    - Validation before any operation
    - Automatic state backup before apply
    - Cost estimation (if infracost is installed)
    - Drift detection
    - Confirmation prompts for destructive operations

EOF
}

# Parse command line arguments
while [[ $# -gt 0 ]]; do
    case $1 in
        -e|--environment)
            ENVIRONMENT="$2"
            shift 2
            ;;
        -a|--action)
            ACTION="$2"
            shift 2
            ;;
        -f|--var-file)
            VAR_FILE="$2"
            shift 2
            ;;
        -t|--target)
            TARGET="$2"
            shift 2
            ;;
        -y|--auto-approve)
            AUTO_APPROVE=true
            shift
            ;;
        -d|--destroy)
            DESTROY=true
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
    echo "Thread Art Generator - Deployment Script"
    echo "========================================"
    echo
    
    print_status "Configuration:"
    echo "  Environment: $ENVIRONMENT"
    echo "  Action: $ACTION"
    echo "  Auto-approve: $AUTO_APPROVE"
    echo "  Destroy: $DESTROY"
    echo "  Target: ${TARGET:-all resources}"
    echo "  Var file: ${VAR_FILE:-default}"
    echo
    
    # Change to environment directory
    # Determine if we're already in the terraform directory or need to navigate to it
    if [ -d "environments/$ENVIRONMENT" ]; then
        # We're in the terraform directory
        local terraform_dir="environments/$ENVIRONMENT"
    elif [ -d "terraform/environments/$ENVIRONMENT" ]; then
        # We're in the project root
        local terraform_dir="terraform/environments/$ENVIRONMENT"
    else
        print_error "Environment directory not found. Looked for:"
        print_error "  - environments/$ENVIRONMENT (from terraform dir)"
        print_error "  - terraform/environments/$ENVIRONMENT (from project root)"
        exit 1
    fi
    
    cd "$terraform_dir"
    
    # Check initialization
    check_terraform_init
    
    # Always validate first
    validate_terraform
    
    # Backup state before destructive operations
    if [ "$ACTION" = "apply" ] || [ "$DESTROY" = true ]; then
        backup_state
    fi
    
    # Execute requested action
    case $ACTION in
        plan)
            terraform_plan
            ;;
        apply)
            terraform_apply
            show_outputs
            ;;
        output)
            show_outputs
            ;;
        check)
            check_drift
            ;;
        summary)
            show_summary
            ;;
        backup)
            backup_state
            ;;
        *)
            print_error "Unknown action: $ACTION"
            usage
            exit 1
            ;;
    esac
    
    print_success "Operation completed successfully!"
}

# Run main function
main "$@"