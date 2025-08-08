# Thread Art Generator - Infrastructure

This directory contains the Terraform infrastructure code for the Thread Art Generator application. The infrastructure is designed to be cost-effective, secure, and production-ready while maintaining a strict $50 monthly budget limit with automatic shutdown protection.

## 🏗️ Architecture Overview

### Core Components

- **Cloud Run Services**: API (internal), Client (public), Worker (internal)
- **Cloud SQL**: PostgreSQL database with private networking
- **Cloud Storage**: Public bucket for CDN content, private bucket for user data
- **Cloud Memorystore Redis**: Session storage and caching
- **Cloud Pub/Sub**: Message queuing (replacing RabbitMQ)
- **Artifact Registry**: Docker image storage
- **Secret Manager**: Secure configuration management
- **VPC Network**: Private networking with Cloud NAT
- **Billing Budget**: $50 hard limit with auto-shutdown

### Security Features

- **Workload Identity Federation**: Secure GitHub Actions authentication
- **Service Accounts**: Least-privilege access for each service
- **IAM Authentication**: Cloud SQL uses service account authentication (no passwords)
- **Private Networking**: Internal-only API and worker services
- **Secret Management**: Application secrets in Google Secret Manager
- **IAM Policies**: Fine-grained access controls

### Cost Optimization

- **Free Tier Resources**: db-f1-micro, basic Redis, minimal storage
- **Auto-scaling**: Scale to zero when not in use
- **Lifecycle Policies**: Automatic cleanup of old images and data
- **Billing Alerts**: Multiple threshold notifications
- **Auto-shutdown**: Automatic service shutdown at 90% budget

## 📁 Directory Structure

```
terraform/
├── environments/           # Environment-specific configurations
│   └── staging/           # Staging environment
│       ├── main.tf        # Main configuration
│       ├── variables.tf   # Variable definitions
│       ├── outputs.tf     # Output definitions
│       └── terraform.tfvars.example
├── modules/               # Reusable Terraform modules
│   ├── artifact-registry/ # Docker image storage
│   ├── billing/          # Budget and cost management
│   ├── cloud-run/        # Container services
│   ├── cloud-sql/        # PostgreSQL database
│   ├── iam/              # Service accounts and permissions
│   ├── networking/       # VPC and network configuration
│   ├── pubsub/           # Message queuing
│   ├── redis/            # Cache and session storage
│   ├── secret-manager/   # Secure configuration
│   └── storage/          # Object storage buckets
├── scripts/              # Deployment and management scripts
│   ├── setup.sh         # Initial infrastructure setup
│   ├── deploy.sh        # Deployment management
│   └── secrets.sh       # Secret management
└── shared/               # Shared configuration
    ├── providers.tf     # Provider configuration
    └── backend.tf       # Remote state backend
```

## 🚀 Quick Start

### Prerequisites

1. **Google Cloud SDK** installed and authenticated
2. **Terraform** >= 1.5 installed
3. **GCP Project** with billing enabled
4. **Firebase Project** created (for authentication)

### Initial Setup

1. **Clone the repository and navigate to terraform directory**:

   ```bash
   cd terraform
   ```

2. **Run the setup script**:

   ```bash
   ./scripts/setup.sh --project-id YOUR_PROJECT_ID
   ```

3. **Configure variables**:

   ```bash
   cd environments/staging
   cp terraform.tfvars.example terraform.tfvars
   # Edit terraform.tfvars with your values
   ```

4. **Initialize application secrets**:

   ```bash
   ../../scripts/secrets.sh --project-id YOUR_PROJECT_ID --action generate
   ```
   
   **Note**: Database authentication uses IAM - no passwords needed.

5. **Deploy infrastructure**:
   ```bash
   ../../scripts/deploy.sh --action apply
   ```

### Configuration

Update `terraform.tfvars` with your specific values:

```hcl
# Required values
project_id = "your-gcp-project-id"
alert_emails = ["your-email@example.com"]
github_repository = "your-github-username/thread-art-generator"
firebase_project_id = "your-firebase-project-id"

# Optional values (defaults shown)
region = "us-central1"
application_name = "thread-art"  # Change for forks
monthly_budget_amount = 50
database_name = "threadartdb"
```

## 🔀 Fork Customization

If you're forking this repository for your own project, follow these steps to customize the infrastructure:

### Key Configuration Changes

1. **Application Name**: Update `application_name` in `terraform.tfvars` to your project name:
   ```hcl
   application_name = "my-awesome-app"  # Affects all resource names
   ```

2. **GitHub Repository**: Update to your fork's repository:
   ```hcl
   github_repository = "your-username/your-repo-name"
   ```

3. **Project Configuration**: Use your own GCP and Firebase projects:
   ```hcl
   project_id = "your-gcp-project-id"
   firebase_project_id = "your-firebase-project-id"
   ```

### Resource Naming Impact

The `application_name` variable affects the naming of all infrastructure resources:

- **Database**: `{application_name}-db-{environment}-{suffix}`
- **Storage Buckets**: `{application_name}-public-{environment}-{suffix}`
- **VPC Network**: `{application_name}-vpc-{environment}`
- **Service Accounts**: `{application_name}-api-sa-{environment}`
- **All other resources** follow similar patterns

### Migration Considerations

- **State Files**: Each fork should use its own Terraform state bucket
- **Resource Conflicts**: Different `application_name` prevents resource naming conflicts
- **Secrets**: Each fork maintains its own secrets in Google Secret Manager
- **CI/CD**: GitHub Actions will authenticate to your specific GCP project

## 🔧 Module Details

### Billing Module (`modules/billing/`)

Implements strict cost controls with automatic shutdown:

- **Budget**: $50 monthly limit with 50%, 80%, 90%, 100% alerts
- **Auto-shutdown**: Cloud Function automatically stops services at 90% budget
- **Notifications**: Email alerts at each threshold
- **Protection**: Prevents bankruptcy through automatic enforcement

### IAM Module (`modules/iam/`)

Manages service accounts and permissions:

- **Service Accounts**: api-sa, client-sa, worker-sa, cicd-sa, migrator-sa
- **Workload Identity**: Secure GitHub Actions authentication
- **Least Privilege**: Minimal required permissions for each service
- **GitHub Integration**: Direct authentication without service account keys

### Networking Module (`modules/networking/`)

Configures secure networking:

- **VPC Network**: Private network for all resources
- **Private Service Access**: Direct connection to Google APIs
- **Cloud NAT**: Outbound internet access for private resources
- **VPC Connector**: Enables Cloud Run to access VPC resources
- **Firewall Rules**: Secure ingress/egress controls

### Cloud Run Module (`modules/cloud-run/`)

Deploys containerized applications:

- **API Service**: Internal-only gRPC/Connect-RPC server
- **Client Service**: Public-facing web application
- **Worker Service**: Background task processing
- **Auto-scaling**: Scale to zero for cost efficiency
- **Health Checks**: Startup and liveness probes
- **Secret Integration**: Environment variables from Secret Manager

### Cloud SQL Module (`modules/cloud-sql/`)

PostgreSQL database with high availability:

- **Instance**: db-f1-micro (free tier) with private IP
- **Authentication**: IAM-based authentication with service account database users
- **Security**: SSL certificates, authorized networks, no password storage
- **Backup**: Configurable backup and point-in-time recovery
- **Cloud SQL Proxy**: Volume-mounted proxy for secure connections
- **Networking**: Private connection via VPC peering

### Storage Module (`modules/storage/`)

Object storage for application data:

- **Public Bucket**: CDN-enabled for static assets and generated images
- **Private Bucket**: Secure storage for user uploads and data
- **Lifecycle Policies**: Automatic cleanup of old data
- **IAM Bindings**: Service-specific access permissions
- **CORS Configuration**: Browser access for uploads

### Redis Module (`modules/redis/`)

Cache and session storage:

- **Instance**: 1GB basic tier (cost-optimized)
- **Network**: Private VPC access only
- **Auth**: Enabled for security
- **Configuration**: Optimized for session storage and caching
- **Maintenance**: Scheduled during low-traffic hours

### Pub/Sub Module (`modules/pubsub/`)

Message queuing replacing RabbitMQ:

- **Topics**: Composition and image processing queues
- **Subscriptions**: Worker subscriptions with retry policies
- **Dead Letter**: Failed message handling
- **IAM**: Publisher and subscriber permissions
- **Durability**: Message retention and acknowledgment policies

### Artifact Registry Module (`modules/artifact-registry/`)

Docker image storage:

- **Repository**: Multi-regional for high availability
- **Cleanup Policies**: Automatic removal of old images
- **IAM**: Pull access for services, push access for CI/CD
- **Cost Control**: Retention policies to minimize storage costs

### Secret Manager Module (`modules/secret-manager/`)

Secure configuration management:

- **Application Secrets**: JWT tokens, API keys, session encryption keys
- **External Service Secrets**: Firebase, email service credentials
- **IAM Bindings**: Service-specific secret access
- **Versioning**: Secret rotation support
- **Encryption**: Google-managed encryption keys
- **No Database Secrets**: Uses IAM authentication instead of passwords

## 🛠️ Management Scripts

### Setup Script (`scripts/setup.sh`)

Initial infrastructure setup:

```bash
./scripts/setup.sh --project-id PROJECT_ID [OPTIONS]

Options:
  -p, --project-id    GCP Project ID (required)
  -r, --region        GCP Region (default: us-central1)
  -e, --environment   Environment (default: staging)
  -b, --bucket        Terraform state bucket name
  -s, --skip-bucket   Skip bucket creation
```

Features:

- Dependency checking (gcloud, terraform)
- GCP API enablement
- Terraform state bucket creation
- Initial Terraform configuration
- tfvars file generation

### Deploy Script (`scripts/deploy.sh`)

Infrastructure deployment and management:

```bash
./scripts/deploy.sh [OPTIONS]

Options:
  -a, --action        Action: plan, apply, output, check, summary
  -e, --environment   Environment (default: staging)
  -y, --auto-approve  Auto-approve apply
  -d, --destroy       Destroy infrastructure
  -t, --target        Target specific resource
```

Features:

- Terraform validation
- State backup before changes
- Cost estimation (if infracost installed)
- Drift detection
- Resource summaries

### Secret Management (`scripts/secrets.sh`)

Secure secret management:

```bash
./scripts/secrets.sh --project-id PROJECT_ID [OPTIONS]

Options:
  -a, --action      Action: list, create, set, get, delete, init, generate
  -s, --secret      Secret name
  -v, --value       Secret value
  -f, --file        Secret file
```

Features:

- Interactive secret creation
- Secure value generation for application secrets
- Batch initialization (excludes database secrets - uses IAM)
- Export to environment files
- Audit and listing

## 🔒 Security Best Practices

### Authentication & Authorization

- **Workload Identity**: GitHub Actions authenticate without service account keys
- **Service Accounts**: Dedicated accounts for each service with minimal permissions
- **IAM Database Authentication**: Cloud SQL uses service account authentication
- **IAM Policies**: Principle of least privilege consistently applied
- **Secret Management**: Application secrets in Google Secret Manager, GCP services use IAM

### Network Security

- **Private Networking**: API and worker services not publicly accessible
- **VPC Isolation**: Resources isolated in dedicated VPC
- **SSL/TLS**: Encryption in transit for all communications
- **Firewall Rules**: Restrictive ingress/egress controls

### Data Protection

- **Encryption**: At-rest encryption for all storage
- **Backup Security**: Encrypted database backups
- **Secret Rotation**: Support for rotating sensitive credentials
- **Audit Logging**: Cloud Audit Logs for all operations

## 💰 Cost Management

### Budget Controls

- **Hard Limit**: $50 monthly budget with automatic enforcement
- **Alert Thresholds**: 50%, 80%, 90%, 100% notifications
- **Auto-shutdown**: Automatic service shutdown at 90% budget
- **Cost Monitoring**: Regular cost analysis and optimization

### Free Tier Optimization

- **Database**: db-f1-micro instance (free tier)
- **Redis**: 1GB basic tier (minimal cost)
- **Cloud Run**: Scale-to-zero with minimal resource allocation
- **Storage**: Lifecycle policies for automatic cleanup
- **Networking**: Free tier VPC and load balancing

### Resource Efficiency

- **Auto-scaling**: Scale down to zero during inactivity
- **Image Cleanup**: Automatic removal of old container images
- **Data Lifecycle**: Automated cleanup of old logs and temporary data
- **Rightsizing**: Appropriately sized resources for staging workload

## 🔄 CI/CD Integration

### GitHub Actions Setup

The infrastructure supports secure CI/CD through Workload Identity Federation:

1. **Authentication**: No service account keys required
2. **Permissions**: Minimal required permissions for deployment
3. **Security**: Identity federation with GitHub OIDC
4. **Artifacts**: Automatic image building and pushing

#### ⚠️ Bootstrap Process - First Deployment Only

**IMPORTANT**: Due to the chicken-and-egg problem with Terraform state storage, the first deployment requires a manual bootstrap step:

##### The Problem
- **Terraform state** is stored in a Google Cloud Storage bucket
- **IAM permissions** for the CI/CD service account are defined in Terraform code
- **To apply Terraform**, the CI/CD service account needs storage permissions
- **But storage permissions** are defined in the Terraform code that hasn't been applied yet!

```
┌─────────────────────────────────────────────────────────────┐
│                    Bootstrap Process                        │
│                                                             │
│  1. Manual Permission Grant (one-time only)                │
│     └─ Grants storage access to CI/CD service account      │
│                                                             │
│  2. Terraform Apply                                         │
│     └─ Creates permanent IAM permissions                   │
│                                                             │
│  3. Remove Manual Permission (optional)                    │
│     └─ Now managed by Terraform                            │
└─────────────────────────────────────────────────────────────┘
```

##### Bootstrap Steps

1. **Grant temporary storage permissions**:
   ```bash
   gcloud projects add-iam-policy-binding YOUR_PROJECT_ID \
     --member="serviceAccount:cicd-sa-staging@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/storage.admin"
   ```

2. **Run Terraform apply** (via GitHub Actions or manually)

3. **Optional cleanup** (permission now managed by Terraform):
   ```bash
   gcloud projects remove-iam-policy-binding YOUR_PROJECT_ID \
     --member="serviceAccount:cicd-sa-staging@YOUR_PROJECT_ID.iam.gserviceaccount.com" \
     --role="roles/storage.admin"
   ```

##### Why This Happens
This is a common **Infrastructure as Code bootstrapping problem**:
- **First deployment**: Manual steps needed for initial permissions
- **Subsequent deployments**: Terraform manages everything automatically
- **Similar examples**: Creating Terraform state bucket, first admin user, initial service accounts

#### Configuration Steps

After deploying the infrastructure, configure GitHub repository variables and secrets:

1. **Get your GCP project information**:
   ```bash
   # Get project ID (if you don't already know it)
   gcloud config get-value project
   
   # Get project number
   gcloud projects describe YOUR_PROJECT_ID --format="value(projectNumber)"
   ```

2. **Add GitHub Repository Variables** in Settings → Secrets and variables → Actions → Variables:
   - `GCP_PROJECT_ID`: Your GCP project ID (e.g., `thread-art-staging-466319`)
   - `GCP_PROJECT_NUMBER`: Your GCP project number (numeric, e.g., `123456789012`)

   **⚠️ Important**: Use **Variables**, not **Secrets** for project ID and number as they're not sensitive and needed in multiple workflow contexts.

3. **Get Terraform outputs** (after first successful deployment):
   ```bash
   cd terraform/environments/staging
   terraform output workload_identity_provider
   terraform output cicd_service_account_email
   ```

4. **Verify GitHub Actions workflow** uses the correct authentication:
   ```yaml
   - name: Authenticate to Google Cloud
     uses: google-github-actions/auth@v2
     with:
       workload_identity_provider: projects/${{ vars.GCP_PROJECT_NUMBER }}/locations/global/workloadIdentityPools/github-staging/providers/github-provider
       service_account: cicd-sa-staging@${{ vars.GCP_PROJECT_ID }}.iam.gserviceaccount.com
   ```

### Deployment Pipeline

```yaml
# Example GitHub Actions workflow
- name: Authenticate to Google Cloud
  uses: google-github-actions/auth@v1
  with:
    workload_identity_provider: ${{ secrets.WIF_PROVIDER }}
    service_account: ${{ secrets.WIF_SERVICE_ACCOUNT }}

- name: Deploy to Cloud Run
  run: |
    gcloud run deploy api --image ${{ env.API_IMAGE }}
```

## 📊 Monitoring & Observability

### Built-in Monitoring

- **Cloud Run Metrics**: Request rate, latency, error rate
- **Cloud SQL Monitoring**: Connection count, CPU, memory usage
- **Redis Metrics**: Memory usage, connection count, hit rate
- **Storage Metrics**: Request rate, bandwidth, storage usage

### Billing Monitoring

- **Budget Alerts**: Proactive notifications at multiple thresholds
- **Cost Breakdown**: Per-service cost analysis
- **Usage Trends**: Historical cost and usage patterns
- **Auto-shutdown**: Automatic protection against budget overruns

### Alerting

- **Email Notifications**: Budget and service health alerts
- **Slack Integration**: Optional team notifications
- **Error Monitoring**: Application error tracking
- **Performance Alerts**: Latency and availability monitoring

## 🚨 Disaster Recovery

### Backup Strategy

- **Database Backups**: Automated daily backups with point-in-time recovery
- **State Backups**: Terraform state versioning in GCS
- **Image Backups**: Container images in Artifact Registry
- **Configuration Backups**: Infrastructure as Code in Git

### Recovery Procedures

1. **Infrastructure Recovery**: Terraform apply from version control
2. **Database Recovery**: Point-in-time restore from backups
3. **Application Recovery**: Deploy from container images
4. **Data Recovery**: Restore from storage backups

### Business Continuity

- **Multi-regional Storage**: Automatic replication for durability
- **Database Redundancy**: Configurable high availability
- **Container Registry**: Multi-regional image storage
- **DNS Failover**: Health check-based routing

## 🔧 Troubleshooting

### Common Issues

1. **Budget Exceeded**: Services automatically shut down at 90% budget

   - **Solution**: Review usage, optimize resources, or increase budget

2. **Permission Denied**: Service account lacks required permissions

   - **Solution**: Check IAM bindings and Workload Identity configuration

3. **Database Connection**: Cannot connect to Cloud SQL

   - **Solution**: Verify VPC connector and private service access

4. **Secret Access**: Cannot read secrets from Secret Manager
   - **Solution**: Check secret names and service account permissions

5. **Database Connection with IAM**: Cannot connect to Cloud SQL with IAM authentication
   - **Solution**: Verify Cloud SQL Auth Proxy volume mount and IAM database users

### Debug Commands

```bash
# Check infrastructure status
./scripts/deploy.sh --action summary

# Verify application secret access (database uses IAM)
./scripts/secrets.sh --project-id PROJECT_ID --action list

# Check for configuration drift
./scripts/deploy.sh --action check

# View terraform outputs
./scripts/deploy.sh --action output
```

### Support Resources

- **GCP Documentation**: https://cloud.google.com/docs
- **Terraform Documentation**: https://terraform.io/docs
- **Project Issues**: GitHub Issues for bug reports
- **Cost Optimization**: GCP Cost Management guides

## 📈 Scaling Considerations

### Horizontal Scaling

- **Cloud Run**: Automatic scaling based on request volume
- **Database**: Connection pooling and read replicas
- **Redis**: Cluster mode for high availability
- **Storage**: Automatic scaling and global distribution

### Vertical Scaling

- **Resource Limits**: Easily adjustable CPU and memory limits
- **Database Tier**: Upgrade from free tier for production
- **Redis Memory**: Increase memory for larger cache
- **Network**: Upgrade VPC connector for higher throughput

### Cost Scaling

- **Budget Adjustment**: Increase monthly budget limit
- **Tier Upgrades**: Move to paid tiers for better performance
- **Resource Optimization**: Right-size resources based on usage
- **Multi-environment**: Separate staging and production budgets

Remember: The infrastructure has a hard $50 budget limit with automatic shutdown to prevent unexpected costs. Monitor usage regularly and optimize resources as needed.
