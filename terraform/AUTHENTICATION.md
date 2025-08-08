# Authentication Architecture

This document explains the authentication mechanisms used in the Thread Art Generator infrastructure, emphasizing the use of GCP's native service account authentication over manual secret management.

## 🔐 Authentication Strategy

### **Service Account Authentication (Primary)**

The infrastructure is designed to use **Google Cloud IAM service accounts** for authentication between GCP services, eliminating the need for passwords and manual secret rotation.

#### **Inter-Service Authentication Flow**

```
[Cloud Run] → [Service Account Token] → [GCP Service] ✅
     ↓
No passwords or API keys needed for GCP services
```

### **What Uses Service Account Authentication**

1. **Cloud Run ↔ Cloud SQL**
   - ✅ **IAM Database Authentication** enabled
   - ✅ **Cloud SQL Auth Proxy** via volume mount
   - ✅ Service accounts have `CLOUD_IAM_SERVICE_ACCOUNT` database users
   - ❌ No database passwords in environment variables

2. **Cloud Run ↔ Cloud Storage**
   - ✅ Service accounts have storage IAM roles
   - ✅ Automatic authentication via metadata service
   - ❌ No storage access keys needed

3. **Cloud Run ↔ Secret Manager**
   - ✅ Service accounts have `secretmanager.secretAccessor` role
   - ✅ Automatic authentication via GCP APIs
   - ❌ No API keys needed

4. **Cloud Run ↔ Pub/Sub**
   - ✅ Service accounts have publisher/subscriber roles
   - ✅ Automatic authentication via GCP client libraries
   - ❌ No connection strings needed

5. **Cloud Run ↔ Redis (Cloud Memorystore)**
   - ✅ Private VPC networking with automatic access
   - ✅ Optional: AUTH string via Secret Manager if needed
   - ❌ No manual credential management

6. **GitHub Actions ↔ GCP**
   - ✅ **Workload Identity Federation**
   - ✅ No service account keys stored in GitHub
   - ✅ OIDC-based authentication

## 🏗️ Implementation Details

### **Cloud SQL IAM Authentication**

```hcl
# Enable IAM authentication
database_flags {
  name  = "cloudsql.iam_authentication"
  value = "on"
}

# Create IAM database users (no passwords)
resource "google_sql_user" "api_iam_user" {
  name     = trimsuffix(var.api_service_account_email, ".gserviceaccount.com")
  instance = google_sql_database_instance.main.name
  type     = "CLOUD_IAM_SERVICE_ACCOUNT"
}
```

**Environment Variables in Cloud Run:**
```bash
POSTGRES_USER=api-sa-staging          # Service account name (no .gserviceaccount.com)
POSTGRES_IAM_AUTH=true               # Enable IAM authentication
POSTGRES_HOST=/cloudsql/connection   # Unix socket via Cloud SQL Proxy
```

**Application Connection:**
```go
// Go application connects using IAM authentication
connStr := fmt.Sprintf("user=%s host=%s dbname=%s sslmode=require", 
    username,    // Service account name from env
    socketPath,  // /cloudsql/<connection-name>
    dbName)
```

### **Cloud SQL Proxy Volume Mount**

```hcl
# Cloud Run service configuration
volumes {
  name = "cloudsql"
  cloud_sql_instance {
    instances = [var.database_connection_name]
  }
}

containers {
  volume_mounts {
    name       = "cloudsql"
    mount_path = "/cloudsql"
  }
}
```

### **Service Account Roles**

```hcl
# API Service Account
resource "google_service_account" "api_sa" {
  account_id = "api-sa-${var.environment}"
}

# Minimal required permissions
resource "google_project_iam_member" "api_sql_client" {
  role   = "roles/cloudsql.client"
  member = "serviceAccount:${google_service_account.api_sa.email}"
}

resource "google_storage_bucket_iam_member" "api_storage_access" {
  role   = "roles/storage.objectAdmin"
  member = "serviceAccount:${google_service_account.api_sa.email}"
}
```

## 🔑 What Still Uses Secrets

### **Application-Level Secrets (Required)**

These secrets are stored in **Google Secret Manager** and accessed via service account authentication:

1. **`token_symmetric_key`** - JWT token signing
2. **`internal_api_key`** - Service-to-service API validation
3. **`cookie_hash_key`** - Session cookie encryption
4. **`cookie_block_key`** - Session cookie encryption

### **External Service Secrets (Manual)**

These are manually managed secrets for external services:

1. **`firebase_web_config`** - Firebase client configuration
2. **`firebase_service_account`** - Firebase admin SDK credentials
3. **`sendinblue_api_key`** - Email service API key

### **Secret Access Pattern**

```hcl
# Secret Manager access via service account
env {
  name = "TOKEN_SYMMETRIC_KEY"
  value_source {
    secret_key_ref {
      secret  = "token-symmetric-key-staging"
      version = "latest"
    }
  }
}
```

## 🔄 Service-to-Service Authentication

### **API ↔ Worker Communication**

Instead of API keys, we can implement **service account token validation**:

```go
// API generates signed token for worker
token, err := generateServiceAccountToken(ctx, workerServiceAccount)

// Worker validates token
claims, err := validateServiceAccountToken(ctx, token)
```

### **Future Enhancement: Remove Internal API Key**

The `internal_api_key` can be eliminated by implementing:

1. **Service Account Token Exchange**
2. **GCP Identity and Access Management**
3. **Workload Identity for service communication**

## 📊 Authentication Summary

| Service Connection | Method | Secrets Required |
|-------------------|--------|------------------|
| Cloud Run → Cloud SQL | IAM Auth + Proxy | ❌ None |
| Cloud Run → Storage | Service Account IAM | ❌ None |
| Cloud Run → Secret Manager | Service Account IAM | ❌ None |
| Cloud Run → Pub/Sub | Service Account IAM | ❌ None |
| Cloud Run → Redis | VPC + Optional AUTH | ⚠️ Optional |
| GitHub Actions → GCP | Workload Identity | ❌ None |
| Application JWT | Secret Manager | ✅ Symmetric Key |
| Session Cookies | Secret Manager | ✅ Hash/Block Keys |
| External APIs | Secret Manager | ✅ API Keys |

## 🛡️ Security Benefits

### **Automatic Token Management**
- ✅ Tokens automatically refreshed
- ✅ Short-lived access tokens (1 hour)
- ✅ No token storage or rotation needed

### **Audit and Monitoring**
- ✅ All access logged in Cloud IAM logs
- ✅ Service account activity tracking
- ✅ No credential leakage in logs

### **Principle of Least Privilege**
- ✅ Each service has minimal required permissions
- ✅ Granular IAM roles per resource
- ✅ No over-privileged access keys

### **Zero Secret Rotation**
- ✅ GCP services require no credential rotation
- ✅ Only application secrets need rotation
- ✅ Reduced operational overhead

## 🔧 Application Integration

### **Go Application Example**

```go
package main

import (
    "context"
    "database/sql"
    "os"
    
    "cloud.google.com/go/cloudsqlconn"
    _ "github.com/lib/pq"
)

func connectToDatabase(ctx context.Context) (*sql.DB, error) {
    // Use IAM authentication
    useIAM := os.Getenv("POSTGRES_IAM_AUTH") == "true"
    user := os.Getenv("POSTGRES_USER")
    
    if useIAM {
        // Connect via Cloud SQL Auth Proxy with IAM
        d, err := cloudsqlconn.NewDialer(ctx)
        if err != nil {
            return nil, err
        }
        
        connStr := fmt.Sprintf("user=%s dbname=%s sslmode=disable", 
            user, os.Getenv("POSTGRES_DB"))
        
        return sql.Open("postgres", connStr)
    }
    
    // Fallback to password authentication (deprecated)
    password := os.Getenv("POSTGRES_PASSWORD")
    connStr := fmt.Sprintf("user=%s password=%s dbname=%s sslmode=require",
        user, password, os.Getenv("POSTGRES_DB"))
    
    return sql.Open("postgres", connStr)
}
```

### **Environment Variables**

```bash
# Database (IAM Authentication)
POSTGRES_USER=api-sa-staging
POSTGRES_DB=threadmachine
POSTGRES_IAM_AUTH=true

# Firebase Storage (Service Account Authentication)
STORAGE_PROVIDER=firebase
STORAGE_PUBLIC_BUCKET=thread-art-public-staging
STORAGE_PRIVATE_BUCKET=thread-art-private-staging

# Pub/Sub (Service Account Authentication)
PUBSUB_PROJECT_ID=thread-art-staging
PUBSUB_TOPIC_PREFIX=thread-art

# Application Secrets (from Secret Manager)
TOKEN_SYMMETRIC_KEY=<from-secret-manager>
INTERNAL_API_KEY=<from-secret-manager>
```

## 🚀 Migration Path

### **Current State**
- ✅ Cloud SQL IAM authentication implemented
- ✅ Service account-based GCP service access
- ✅ Workload Identity Federation for CI/CD
- ⚠️ Some application secrets still in Secret Manager

### **Future Optimizations**

1. **Phase 1: Remove Internal API Key**
   - Implement service account token validation
   - Replace API key with signed tokens

2. **Phase 2: Enhance Service Communication**
   - Use GCP Identity-Aware Proxy for internal services
   - Implement mTLS for service-to-service communication

3. **Phase 3: External Secret Optimization**
   - Evaluate external service authentication methods
   - Implement OAuth 2.0 where possible

## 📝 Operational Benefits

### **Simplified Ops**
- ❌ No database password rotation needed
- ❌ No storage access key management
- ❌ No service account key files
- ✅ Automatic credential lifecycle management

### **Enhanced Security**
- 🔒 All database access is audited via IAM
- 🔒 No passwords in environment variables
- 🔒 Short-lived, automatically rotated tokens
- 🔒 Granular permission control

### **Cost Efficiency**
- 💰 No external credential management systems needed
- 💰 Built-in GCP IAM at no additional cost
- 💰 Reduced operational overhead

This authentication architecture provides maximum security with minimal operational overhead, leveraging GCP's native identity and access management capabilities.