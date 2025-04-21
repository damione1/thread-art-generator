# Auth0 Integration Guide

This guide explains how to set up Auth0 authentication with the Thread Art Generator application.

> **For local development with HTTPS**: Please refer to [Local Development with Auth0](./docs/AUTH0_LOCAL_DEV.md) for setting up Auth0 with tag.local domain.

## 1. Create an Auth0 Account

If you don't already have one, create an account at [Auth0](https://auth0.com/).

## 2. Create a New API

1. In the Auth0 dashboard, go to **APIs** and click **Create API**
2. Set a name (e.g., "Thread Art Generator API")
3. Set an identifier (e.g., `https://api.threadart.com`)
4. Select RS256 as the signing algorithm
5. Click **Create**

## 3. Configure the API

1. In your API settings, go to the **Settings** tab
2. Make sure **Token Expiration** is set to an appropriate time (e.g., 86400 seconds/24 hours)
3. Enable **Allow Offline Access** to get refresh tokens

## 4. Create an Application

1. Go to **Applications** and click **Create Application**
2. Select **Single Page Application** for a React/NextJS frontend
3. Set a name (e.g., "Thread Art Generator Web")
4. Click **Create**

## 5. Configure the Application

1. In your application settings, set these values:

   - **Allowed Callback URLs**: `http://localhost:3000/callback, https://your-production-url.com/callback`
   - **Allowed Logout URLs**: `http://localhost:3000, https://your-production-url.com`
   - **Allowed Web Origins**: `http://localhost:3000, https://your-production-url.com`
   - **Allowed Origins (CORS)**: `http://localhost:3000, https://your-production-url.com`

2. Save changes

## 6. Create Rules for User Registration

Create an Auth0 Action to call your API after user registration:

1. Go to **Actions** > **Flows** > **Login**
2. Add a new action named "Create User in Database"
3. Use this code (customize as needed):

```javascript
exports.onExecutePostLogin = async (event, api) => {
  // Only run for new signups
  if (event.stats.logins_count > 1) return;

  const axios = require("axios");

  try {
    // Get a management API token
    const domain = event.secrets.AUTH0_DOMAIN;
    const clientId = event.secrets.AUTH0_CLIENT_ID;
    const clientSecret = event.secrets.AUTH0_CLIENT_SECRET;

    const tokenResponse = await axios.post(`https://${domain}/oauth/token`, {
      client_id: clientId,
      client_secret: clientSecret,
      audience: `https://${domain}/api/v2/`,
      grant_type: "client_credentials",
    });

    const managementToken = tokenResponse.data.access_token;

    // Call your API to create a user
    await axios.post(
      "https://your-api.com/v1/users",
      {
        user: {
          auth0_id: event.user.user_id,
          email: event.user.email,
          first_name: event.user.given_name || "",
          last_name: event.user.family_name || "",
        },
      },
      {
        headers: {
          Authorization: `Bearer ${managementToken}`,
          "Content-Type": "application/json",
        },
      }
    );

    console.log("User successfully created in application database");
  } catch (error) {
    console.error("Error creating user in database:", error);
    // Consider whether to block login if user creation fails
    // api.access.deny('Failed to create user account');
  }
};
```

4. Add secrets for AUTH0_DOMAIN, AUTH0_CLIENT_ID, and AUTH0_CLIENT_SECRET
5. Deploy the action and add it to the Login flow

## 7. Configure Environment Variables

Add these variables to your application's environment:

```
AUTH0_DOMAIN=your-tenant.auth0.com
AUTH0_AUDIENCE=https://api.threadart.com
AUTH0_CLIENT_ID=your-client-id
AUTH0_CLIENT_SECRET=your-client-secret
```

## 8. Update Your User Model

Make sure your user model has an `auth0_id` field to link Auth0 users with your application users.

## 9. NextJS Frontend Integration

In your NextJS application, install the Auth0 SDK:

```bash
npm install @auth0/auth0-react
```

Create an Auth0 provider in your `_app.tsx`:

```tsx
import { Auth0Provider } from "@auth0/auth0-react";

function MyApp({ Component, pageProps }) {
  return (
    <Auth0Provider
      domain={process.env.NEXT_PUBLIC_AUTH0_DOMAIN}
      clientId={process.env.NEXT_PUBLIC_AUTH0_CLIENT_ID}
      authorizationParams={{
        redirect_uri:
          typeof window !== "undefined" ? window.location.origin : "",
        audience: process.env.NEXT_PUBLIC_AUTH0_AUDIENCE,
      }}
    >
      <Component {...pageProps} />
    </Auth0Provider>
  );
}

export default MyApp;
```

## 10. Making Authenticated API Calls

Use this pattern to make authenticated calls:

```tsx
import { useAuth0 } from '@auth0/auth0-react';
import { MyServiceClient } from '../generated/service_grpc_web_pb';

function MyComponent() {
  const { getAccessTokenSilently, isAuthenticated } = useAuth0();

  async function callApi() {
    if (!isAuthenticated) return;

    const token = await getAccessTokenSilently();
    const client = new MyServiceClient('https://your-api.com');

    // Set up request
    const request = new SomeRequest();

    // Add token to metadata
    const metadata = {'Authorization': `Bearer ${token}`};

    // Make gRPC call
    client.someMethod(request, metadata, (err, response) => {
      if (err) {
        console.error(err);
        return;
      }
      console.log(response);
    });
  }

  return (
    // Your component JSX
  );
}
```

## 11. Testing

1. Start your application
2. Visit your frontend
3. Test login/signup with Auth0
4. Confirm the user is created in your database
5. Test authenticated API calls
