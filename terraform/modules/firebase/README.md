# Firebase Module

This Terraform module provisions Firebase resources including:

- Firebase project configuration
- Firebase Authentication
- Cloud Functions for user synchronization
- Web app configuration

## Resources Created

### Firebase Core
- `google_firebase_project` - Configures the GCP project as a Firebase project
- `google_firebase_web_app` - Creates a Firebase web application
- `google_identity_platform_config` - Configures Firebase Authentication

### Cloud Functions
- `google_cloudfunctions2_function.sync_user_on_create` - Syncs Firebase users to backend database
- `google_cloudfunctions2_function.health_check` - Health check endpoint for monitoring
- `google_storage_bucket.functions_source` - Storage for Cloud Functions source code

### Authentication Providers (Optional)
- `google_identity_platform_default_supported_idp_config` - Google OAuth provider
- `google_identity_platform_default_supported_idp_config` - GitHub OAuth provider

## Prerequisites

1. **Functions Source Code**: The module expects Firebase Functions source code to be present in the directory specified by `functions_source_dir` (default: `../../../functions`)

2. **Backend API**: The backend API service must be deployed and accessible at the URL specified in `backend_api_url`

3. **Internal API Key**: A secure internal API key must be provided for Cloud Functions to authenticate with the backend

## Usage

```hcl
module "firebase" {
  source = "../../modules/firebase"

  project_id       = var.project_id
  environment      = var.environment
  region           = var.region
  application_name = var.application_name

  # Firebase Authentication configuration
  authorized_domains = [
    "localhost",
    "${var.project_id}.web.app",
    "${var.project_id}.firebaseapp.com"
  ]

  # Backend integration
  backend_api_url  = "https://your-api-service-url"
  internal_api_key = var.internal_api_key

  # Service account integration
  api_service_account_email = module.iam.api_service_account_email
}
```

## Required APIs

The module automatically enables the following Google Cloud APIs:
- Firebase API (`firebase.googleapis.com`)
- Firebase Authentication API (`firebaseauth.googleapis.com`)
- Cloud Functions API (`cloudfunctions.googleapis.com`)
- Cloud Build API (`cloudbuild.googleapis.com`)
- Artifact Registry API (`artifactregistry.googleapis.com`)
- Cloud Run API (`run.googleapis.com`)
- Eventarc API (`eventarc.googleapis.com`)
- Pub/Sub API (`pubsub.googleapis.com`)

## Environment Variables

The Cloud Functions will be configured with the following environment variables:
- `NODE_ENV`: Set to the value of `var.environment`
- `BACKEND_URL`: Set to the value of `var.backend_api_url`
- `INTERNAL_API_KEY`: Set to the value of `var.internal_api_key`

## OAuth Providers

OAuth providers (Google, GitHub) are disabled by default for cost optimization. To enable:

```hcl
module "firebase" {
  # ... other configuration ...
  
  enable_google_provider = true
  google_oauth_client_id = var.google_oauth_client_id
  google_oauth_client_secret = var.google_oauth_client_secret
  
  enable_github_provider = true
  github_oauth_client_id = var.github_oauth_client_id
  github_oauth_client_secret = var.github_oauth_client_secret
}
```

## Outputs

The module provides the following outputs:
- `firebase_project_id`: The Firebase project ID
- `firebase_config`: Complete Firebase configuration for client applications
- `firebase_auth_domain`: Firebase Auth domain
- `sync_user_function_name`: Name of the user sync Cloud Function
- `health_check_function_url`: URL for the health check endpoint

## Security Considerations

1. **Internal API Key**: Store the internal API key securely (e.g., in Google Secret Manager or as a GitHub secret)
2. **Function IAM**: Cloud Functions use least-privilege IAM roles
3. **Authorized Domains**: Configure only trusted domains for Firebase Auth
4. **OAuth Secrets**: Store OAuth client secrets securely and mark them as sensitive

## Cost Optimization

For staging environments, the module is configured with:
- Minimal memory allocation for Cloud Functions (256Mi)
- Low max instance count (5)
- Short timeout (30 seconds)
- Disabled OAuth providers
- Automatic source code cleanup (7 days retention)

## Troubleshooting

### Common Issues

1. **Functions deployment fails**: Ensure the `functions_source_dir` contains valid Node.js code with required dependencies
2. **User sync fails**: Verify that `backend_api_url` is accessible and `internal_api_key` is correct
3. **Authentication not working**: Check that authorized domains include your frontend domain

### Debugging

Check Cloud Function logs:
```bash
gcloud functions logs read --region=us-central1 thread-art-sync-user-staging
```

Test health check endpoint:
```bash
curl https://REGION-PROJECT_ID.cloudfunctions.net/thread-art-functions-health-staging
```

## Dependencies

This module should be deployed after:
- IAM module (for service accounts)
- Secrets module (for API keys)

And before:
- Cloud Run module (which depends on Firebase project ID)