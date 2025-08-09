# Firebase Storage Security Configuration

## Overview

This project implements a dual-security approach for Firebase Storage:

- **Local Development**: Permissive rules (`storage.rules`) for Firebase emulator compatibility
- **Production**: Extremely strict rules (`storage.production.rules`) that rely on signed URLs only

## Security Architecture

### Production Security Model

1. **Signed URLs Only**: All uploads must go through the API's signed URL generation
2. **Path-Based Access**: Public read access only for `users/{uid}/arts/{art_id}`
3. **Authentication Required**: All other reads require Firebase Auth with UID matching
4. **Zero Direct Access**: No direct uploads or unauthorized reads allowed

### Local Development Model

- Uses permissive rules because Firebase emulator doesn't support signed URLs
- Allows direct uploads and reads for development convenience
- Should NEVER be deployed to production

## File Structure

```
├── storage.rules              # Local development (emulator only)
├── storage.production.rules   # Production security rules
└── terraform/modules/firebase/
    └── main.tf               # Deploys production rules
```

## Production Rules Breakdown

```javascript
// PUBLIC READ: Allow public viewing of art pieces
match /users/{uid}/arts/{art_id} {
  allow read: if true;        // Public art display
  allow write: if false;      // No direct writes
}

// AUTHENTICATED READ: Users can read their own private files  
match /users/{uid}/{allPaths=**} {
  allow read: if request.auth != null && request.auth.uid == uid;
  allow write: if false;      // No direct writes
}

// EVERYTHING ELSE: Completely locked down
match /{allPaths=**} {
  allow read: if false;       // No access
  allow write: if false;      // No access  
}
```

## Deployment Instructions

### Automated Deployment (Terraform)

The Terraform configuration automatically:
1. Uploads production rules file to Cloud Storage
2. Creates deployment reminder for manual rules deployment
3. Configures environment-specific IAM permissions

### Manual Production Deployment

**Required**: Firebase Storage rules must be manually deployed to production:

```bash
# Deploy production storage rules
firebase deploy --only storage --project your-production-project-id
```

**Important**: Ensure `storage.production.rules` is configured in your `firebase.json`:

```json
{
  "storage": {
    "rules": "storage.production.rules"
  }
}
```

## Security Validations

### Upload Security
✅ Only API-generated signed URLs can upload
✅ 1-minute expiration on upload URLs
✅ Content type validation (images only)
✅ No direct upload paths in production

### Read Security  
✅ Public read only for `users/{uid}/arts/{art_id}` 
✅ Private files require authentication + UID match
✅ No IAM public read bindings in production
✅ All access controlled by Firebase Storage rules

### Environment Isolation
✅ Different rules for development vs production
✅ Emulator uses permissive rules for dev experience  
✅ Production enforces signed URL requirement
✅ IAM permissions environment-specific

## Testing Security

### Production Tests
1. **Unauthorized Upload**: Should fail without signed URL
2. **Cross-User Access**: User A cannot read User B's private files
3. **Public Art Access**: Anyone can read `users/{uid}/arts/{art_id}`
4. **Authenticated Private**: User can read own files with valid token

### Development Tests
1. **Emulator Upload**: Direct uploads work for development
2. **Local Access**: Direct file access works without signed URLs

## Troubleshooting

### Common Issues

**Rules Not Applied**: 
- Check Firebase project ID matches Terraform project
- Manually deploy rules: `firebase deploy --only storage`
- Verify `firebase.json` points to correct rules file

**Emulator Upload Fails**:
- Ensure using development `storage.rules` (permissive)
- Check emulator host configuration
- Verify network connectivity

**Production Upload Blocked**:
- ✅ Expected behavior - must use signed URLs
- Check signed URL generation in API
- Verify content type and expiration

## Security Checklist

- [ ] Production uses `storage.production.rules`
- [ ] Development uses `storage.rules` 
- [ ] No public IAM bindings in production
- [ ] Signed URLs have 1-minute expiration
- [ ] Content type validation enabled
- [ ] Firebase Auth integration working
- [ ] Path-based access controls tested
- [ ] Cross-user access blocked