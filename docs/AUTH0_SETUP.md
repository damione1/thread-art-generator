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
2. Select **Regular Web Application** for a Go+HTMX frontend
3. Set a name (e.g., "Thread Art Generator Web")
4. Click **Create**

## 5. Configure the Application

1. In your application settings, set these values:

   - **Allowed Callback URLs**: `http://localhost:8080/callback, https://your-production-url.com/callback`
   - **Allowed Logout URLs**: `http://localhost:8080, https://your-production-url.com`
   - **Allowed Web Origins**: `http://localhost:8080, https://your-production-url.com`
   - **Allowed Origins (CORS)**: `http://localhost:8080, https://your-production-url.com`

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
AUTH0_CALLBACK_URL=http://localhost:8080/callback
AUTH0_LOGOUT_URL=http://localhost:8080
```

## 8. Update Your User Model

Make sure your user model has an `auth0_id` field to link Auth0 users with your application users.

## 9. Go Backend Integration

In your Go frontend application, install the Auth0 Go SDK:

```bash
go get github.com/coreos/go-oidc/v3/oidc
go get golang.org/x/oauth2
go get github.com/gorilla/sessions
```

Create an Auth0 authenticator in your Go code:

```go
package auth

import (
	"context"
	"errors"
	"net/http"
	"os"

	"github.com/coreos/go-oidc/v3/oidc"
	"github.com/gorilla/sessions"
	"golang.org/x/oauth2"
)

type Authenticator struct {
	Provider     *oidc.Provider
	Config       oauth2.Config
	SessionStore sessions.Store
}

func NewAuthenticator() (*Authenticator, error) {
	provider, err := oidc.NewProvider(
		context.Background(),
		"https://"+os.Getenv("AUTH0_DOMAIN")+"/",
	)
	if err != nil {
		return nil, err
	}

	conf := oauth2.Config{
		ClientID:     os.Getenv("AUTH0_CLIENT_ID"),
		ClientSecret: os.Getenv("AUTH0_CLIENT_SECRET"),
		RedirectURL:  os.Getenv("AUTH0_CALLBACK_URL"),
		Endpoint:     provider.Endpoint(),
		Scopes:       []string{oidc.ScopeOpenID, "profile", "email"},
	}

	// Create session store (use Redis in production)
	sessionStore := sessions.NewCookieStore([]byte(os.Getenv("SESSION_KEY")))
	sessionStore.Options = &sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 1 week
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") != "development",
	}

	return &Authenticator{
		Provider:     provider,
		Config:       conf,
		SessionStore: sessionStore,
	}, nil
}

// Login initiates the authentication flow
func (a *Authenticator) Login(w http.ResponseWriter, r *http.Request) {
	state := generateRandomState()
	session, _ := a.SessionStore.Get(r, "auth-session")
	session.Values["state"] = state
	session.Save(r, w)

	http.Redirect(w, r, a.Config.AuthCodeURL(state), http.StatusTemporaryRedirect)
}

// Callback handles the OAuth2 callback
func (a *Authenticator) Callback(w http.ResponseWriter, r *http.Request) {
	session, _ := a.SessionStore.Get(r, "auth-session")

	// Verify state
	if r.URL.Query().Get("state") != session.Values["state"] {
		http.Error(w, "Invalid state parameter", http.StatusBadRequest)
		return
	}

	// Exchange code for token
	token, err := a.Config.Exchange(context.Background(), r.URL.Query().Get("code"))
	if err != nil {
		http.Error(w, "Failed to exchange token", http.StatusInternalServerError)
		return
	}

	// Get user info
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		http.Error(w, "No id_token field in oauth2 token", http.StatusInternalServerError)
		return
	}

	// Store tokens in session
	session.Values["access_token"] = token.AccessToken
	session.Values["id_token"] = rawIDToken
	session.Save(r, w)

	// Redirect to home page
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
```

## 10. Making Authenticated API Calls

Use this pattern to make authenticated calls to your backend:

```go
import (
	"context"
	"github.com/bufbuild/connect-go"
)

// Handler that makes an authenticated call to the backend
func (h *Handler) SomeAuthenticatedEndpoint(w http.ResponseWriter, r *http.Request) {
	// Get the session
	session, _ := h.Auth.SessionStore.Get(r, "auth-session")

	// Check if the user is authenticated
	if session.Values["access_token"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get the access token from the session
	token := session.Values["access_token"].(string)

	// Create a Connect client with the access token
	client := someservice.NewClient(
		h.ConnectClient,
		connect.WithInterceptors(
			connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
				return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
					// Add the auth header
					req.Header().Set("Authorization", "Bearer "+token)
					return next(ctx, req)
				}
			}),
		),
	)

	// Make the API call
	resp, err := client.SomeMethod(r.Context(), connect.NewRequest(&someservice.SomeRequest{
		// Request parameters
	}))

	if err != nil {
		// Handle error
		http.Error(w, "API call failed", http.StatusInternalServerError)
		return
	}

	// Process the response and render the template
	tmplData := map[string]interface{}{
		"Data": resp.Msg,
	}

	h.Templates.ExecuteTemplate(w, "some-template.html", tmplData)
}
```

## 11. Session Management with Redis

For production environments, implement Redis-based session storage:

```bash
go get github.com/gorilla/sessions
go get github.com/rbcervilla/redisstore/v8
```

```go
package auth

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/gorilla/sessions"
	"github.com/rbcervilla/redisstore/v8"
)

func NewRedisSessionStore() (sessions.Store, error) {
	// Create Redis client
	client := redis.NewClient(&redis.Options{
		Addr:     os.Getenv("REDIS_ADDR"),
		Password: os.Getenv("REDIS_PASSWORD"),
		DB:       0,
	})

	// Create Redis store
	store, err := redisstore.NewRedisStore(context.Background(), client)
	if err != nil {
		return nil, err
	}

	// Configure store
	store.Options(sessions.Options{
		Path:     "/",
		MaxAge:   86400 * 7, // 1 week
		HttpOnly: true,
		Secure:   os.Getenv("ENVIRONMENT") != "development",
	})

	return store, nil
}
```

## 12. Testing

1. Start your application
2. Visit your frontend
3. Click the login button
4. Test the authentication flow with Auth0
5. Confirm the user is created in your database
6. Test authenticated API calls
