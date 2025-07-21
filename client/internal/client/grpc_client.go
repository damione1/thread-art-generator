package client

import (
	"context"
	"fmt"
	"net/http"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pb/pbconnect"

	"connectrpc.com/connect"
)

// APIClient provides a wrapper for the API client with Firebase authentication
// Implements auth.APIClient interface
type APIClient struct {
	baseURL        string
	sessionManager *auth.SCSSessionManager
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

// CheckSessionToken checks if a session has a valid Firebase ID token
func (c *APIClient) CheckSessionToken(r *http.Request) error {
	idToken := c.sessionManager.GetIDToken(r)
	if idToken == "" {
		return fmt.Errorf("session has no Firebase ID token")
	}

	// Log token info for debugging
	fmt.Printf("Session contains Firebase ID token: %s...%s\n",
		idToken[:10],
		idToken[len(idToken)-10:])

	return nil
}

// NewAPIClient creates a new API client using Connect-RPC with Firebase authentication
func NewAPIClient(baseURL string, sessionManager *auth.SCSSessionManager) *APIClient {
	httpClient := &http.Client{
		Transport: &FirebaseAuthTransport{
			SessionManager: sessionManager,
			Base:           http.DefaultTransport,
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
		// Get Firebase ID token from SCS session
		idToken := c.sessionManager.GetIDToken(req)
		if idToken == "" {
			return nil, fmt.Errorf("not authenticated: no Firebase ID token")
		}
		// Create a context with the token for the FirebaseAuthTransport
		ctxWithSession = context.WithValue(ctx, "firebase_token", idToken)
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
		return nil, fmt.Errorf("no session found: %w", err)
	}

	// Create a user from Firebase session data
	user := &User{
		ID:        session.UserInfo.ID,
		FirstName: session.UserInfo.FirstName,
		LastName:  session.UserInfo.LastName,
		Email:     session.UserInfo.Email,
		Avatar:    session.UserInfo.Picture,
	}

	return user, nil
}

