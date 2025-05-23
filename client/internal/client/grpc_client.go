package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"

	"github.com/bufbuild/connect-go"
)

// APIClient provides a wrapper for the API client with authentication
// Implements auth.APIClient interface
type APIClient struct {
	baseURL        string
	sessionManager *auth.SessionManager
	httpClient     *http.Client
	connectClient  pbconnect.ArtGeneratorServiceClient
}

// Ensure APIClient implements auth.APIClient interface
var _ auth.APIClient = (*APIClient)(nil)

// User represents the data returned from the API
// This implements auth.APIUser for easier type conversion
type User struct {
	ID        string `json:"id"`
	FirstName string `json:"first_name"`
	LastName  string `json:"last_name"`
	Email     string `json:"email"`
	Avatar    string `json:"avatar"`
}

// ToAPIUser converts our User to an auth.APIUser
func (u *User) ToAPIUser() *auth.APIUser {
	return &auth.APIUser{
		ID:        u.ID,
		FirstName: u.FirstName,
		LastName:  u.LastName,
		Email:     u.Email,
		Avatar:    u.Avatar,
	}
}

// CheckSessionToken checks if a session has a valid Auth0 token
func (c *APIClient) CheckSessionToken(r *http.Request) error {
	session, err := c.sessionManager.GetSession(r)
	if err != nil {
		return fmt.Errorf("not authenticated: %w", err)
	}

	// Check if the token exists and has a reasonable format
	if session.AccessToken == "" {
		return fmt.Errorf("session has no access token")
	}

	// Log token info for debugging
	fmt.Printf("Session for user %s contains token: %s...%s\n",
		session.UserID,
		session.AccessToken[:10],
		session.AccessToken[len(session.AccessToken)-10:])

	return nil
}

// NewAPIClient creates a new API client using Connect-RPC
func NewAPIClient(baseURL string, sessionManager *auth.SessionManager) *APIClient {
	httpClient := &http.Client{
		Transport: &authTransport{
			sessionManager: sessionManager,
			base:           http.DefaultTransport,
		},
	}

	connectClient := pbconnect.NewArtGeneratorServiceClient(
		httpClient,
		baseURL,
	)

	return &APIClient{
		baseURL:        baseURL,
		sessionManager: sessionManager,
		httpClient:     httpClient,
		connectClient:  connectClient,
	}
}

// GetCurrentUser fetches the current user from the API using Connect-RPC
func (c *APIClient) GetCurrentUser(ctx context.Context, req *http.Request) (*auth.APIUser, error) {
	var ctxWithSession context.Context

	// Check if we already have a session in the context
	if ctx.Value("session") != nil {
		ctxWithSession = ctx // Use the context as is
	} else if req != nil {
		// Get session from request if provided
		session, err := c.sessionManager.GetSession(req)
		if err != nil {
			return nil, fmt.Errorf("not authenticated: %w", err)
		}
		// Create a context with the session for the authTransport
		ctxWithSession = context.WithValue(ctx, "session", session)
	} else {
		return nil, fmt.Errorf("neither context contains session nor request provided")
	}

	// Create the Connect request
	connectReq := connect.NewRequest(&pb.GetCurrentUserRequest{})

	// Call the Connect-RPC endpoint
	resp, err := c.connectClient.GetCurrentUser(ctxWithSession, connectReq)
	if err != nil {
		return nil, fmt.Errorf("failed to get current user: %w", err)
	}

	// Convert the response to our APIUser type
	user := &auth.APIUser{
		ID:        resp.Msg.GetName(),
		FirstName: resp.Msg.GetFirstName(),
		LastName:  resp.Msg.GetLastName(),
		Email:     resp.Msg.GetEmail(),
		Avatar:    resp.Msg.GetAvatar(),
	}

	return user, nil
}

// GetInternalUser fetches the current user from the API and returns our internal User type
// This is for internal use when we need our own User type
func (c *APIClient) GetInternalUser(ctx context.Context, req *http.Request) (*User, error) {
	apiUser, err := c.GetCurrentUser(ctx, req)
	if err != nil {
		return nil, err
	}

	return &User{
		ID:        apiUser.ID,
		FirstName: apiUser.FirstName,
		LastName:  apiUser.LastName,
		Email:     apiUser.Email,
		Avatar:    apiUser.Avatar,
	}, nil
}

// For simplicity, we'll implement a mock version that just returns session data
func (c *APIClient) GetCurrentUserMock(req *http.Request) (*User, error) {
	session, err := c.sessionManager.GetSession(req)
	if err != nil {
		return nil, err
	}

	// Create a user from session data
	user := &User{
		ID:        session.UserID,
		FirstName: session.UserInfo.FirstName,
		LastName:  session.UserInfo.LastName,
		Email:     session.UserInfo.Email,
		Avatar:    session.UserInfo.Picture,
	}

	return user, nil
}

// authTransport is an http.RoundTripper that adds auth headers from the session
type authTransport struct {
	sessionManager *auth.SessionManager
	base           http.RoundTripper
}

// RoundTrip implements http.RoundTripper
func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// First check for token directly in the context (added by middleware)
	token, ok := auth.TokenFromContext(req.Context())
	if ok && token != "" {
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", token))
		fmt.Println("Auth transport: Using token from context")
	} else if req.Context().Value("session") != nil {
		// Fall back to session from context if available
		session := req.Context().Value("session").(*auth.SessionData)
		// Forward the raw Auth0 token directly - this ensures proper claims extraction on the API side
		req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", session.AccessToken))
	} else {
		fmt.Println("Auth transport: No token or session found in context")
	}

	// Add Origin header to prevent CORS errors
	req.Header.Set("Origin", "http://localhost:8080") // Set to your client's origin

	return t.base.RoundTrip(req)
}
