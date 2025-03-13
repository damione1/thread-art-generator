# Auth0 Setup for Local Development

This guide explains how to set up Auth0 for local development with HTTPS using Traefik and mkcert.

## Prerequisites

- [mkcert](https://github.com/FiloSottile/mkcert) installed (or use the automated setup)
- [Tilt](https://tilt.dev/) installed
- Docker installed

## Setup Steps

### 1. Generate Local Certificates

Run the setup script to generate local certificates and configure your hosts file:

```bash
./setup_mkcert.sh
```

Or use Tilt:

```bash
tilt trigger setup-mkcert
```

This will:

- Install mkcert if not already installed
- Generate certificates for tag.local
- Add tag.local to your hosts file (pointing to 127.0.0.1)

### 2. Start the Development Environment

```bash
tilt up
```

This will:

- Build and start all services including Traefik
- Configure HTTPS for tag.local

### 3. Auth0 Configuration

1. Log in to your [Auth0 Dashboard](https://manage.auth0.com/)
2. Navigate to Applications > Applications
3. Select your application or create a new one
4. Under "Application URIs", add the following:
   - Allowed Callback URLs: `https://tag.local/callback`
   - Allowed Logout URLs: `https://tag.local`
   - Allowed Web Origins: `https://tag.local`
5. Save changes

### 4. Update .env File

Make sure your local .env file has the correct Auth0 configuration:

```
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_AUDIENCE=https://api.your-domain.com
AUTH0_CLIENT_ID=your-client-id
AUTH0_CLIENT_SECRET=your-client-secret
FRONTEND_URL=https://tag.local
```

## Troubleshooting

### Certificate Issues

If you're seeing certificate warnings:

1. Check if the certificate is installed in your system trust store:

   ```bash
   mkcert -CAROOT
   ```

2. Reinstall the certificate:

   ```bash
   mkcert -install
   ```

3. Regenerate certificates:
   ```bash
   rm -rf ./certs
   ./setup_mkcert.sh
   ```

### Auth0 Login Issues

If you're having trouble with Auth0 login:

1. Verify the allowed callback URLs in Auth0 dashboard
2. Check the browser console for any CORS errors
3. Ensure your .env configuration matches Auth0 settings
4. Try clearing browser cookies and cache

### Traefik Dashboard

The Traefik dashboard is available at http://localhost:8080/dashboard/ where you can:

- Verify routes are correctly configured
- Check certificate status
- View incoming requests

## Team Development

When working with a team, each developer needs to:

1. Run the setup script once to generate local certificates
2. Make sure tag.local is added to their hosts file
3. Configure the Auth0 application to include their callback URL

For applications with frequent team changes, consider creating a dedicated Auth0 application for development that includes all possible callback URLs.
