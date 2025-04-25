# Thread Art Generator - Connect-RPC Client Implementation Guidelines for Go

This document provides guidelines for implementing and using Connect-RPC clients in the Go+HTMX frontend of the Thread Art Generator application.

## Architecture Overview

Our system uses:

1. Go+HTMX frontend connecting to a gRPC backend via Connect-RPC
2. Server-side rendered templates with progressive enhancement
3. Session-based authentication with Auth0
4. Redis for distributed session storage

The request flow is as follows:

- Browser makes HTTP requests to the Go frontend
- Go frontend makes Connect-RPC calls to backend services
- Results are rendered as HTML and returned to the browser
- HTMX handles partial updates without full page refreshes

## Core Architecture Components

### 1. Connect-RPC Client Setup

```go
package client

import (
	"context"
	"net/http"
	"time"

	"github.com/bufbuild/connect-go"
	"github.com/your-org/thread-art-generator/core/pb/artgenerator/v1/artgeneratorv1connect"
)

// ClientConfig holds configuration for Connect clients
type ClientConfig struct {
	BaseURL         string
	Timeout         time.Duration
	RetryMax        int
	RetryDelay      time.Duration
	RetryDelayMax   time.Duration
}

// NewArtGeneratorClient creates a new Connect client for the art generator service
func NewArtGeneratorClient(config ClientConfig) artgeneratorv1connect.ArtGeneratorServiceClient {
	// Create HTTP client with timeout
	httpClient := &http.Client{
		Timeout: config.Timeout,
	}

	// Create Connect client with interceptors
	client := artgeneratorv1connect.NewArtGeneratorServiceClient(
		httpClient,
		config.BaseURL,
		connect.WithInterceptors(
			NewAuthInterceptor(),
			NewRetryInterceptor(config),
			NewLoggingInterceptor(),
		),
	)

	return client
}
```

### 2. Authentication Interceptor

```go
// NewAuthInterceptor creates an interceptor that adds auth headers to requests
func NewAuthInterceptor() connect.UnaryInterceptorFunc {
	return connect.UnaryInterceptorFunc(func(next connect.UnaryFunc) connect.UnaryFunc {
		return func(ctx context.Context, req connect.AnyRequest) (connect.AnyResponse, error) {
			// Get token from context (if available)
			if token, ok := auth.TokenFromContext(ctx); ok {
				req.Header().Set("Authorization", "Bearer "+token)
			}

			return next(ctx, req)
		}
	})
}

// WithToken adds an auth token to a context
func WithToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, authTokenKey, token)
}

// TokenFromContext extracts an auth token from a context
func TokenFromContext(ctx context.Context) (string, bool) {
	token, ok := ctx.Value(authTokenKey).(string)
	return token, ok
}
```

### 3. Error Handling

```go
// HandleConnectError processes Connect-RPC errors into user-friendly messages
func HandleConnectError(err error) (FieldErrors, string) {
	fieldErrors := make(FieldErrors)
	var generalError string

	if connectErr, ok := err.(*connect.Error); ok {
		switch connectErr.Code() {
		case connect.CodeUnauthenticated:
			generalError = "Authentication required. Please log in."
		case connect.CodePermissionDenied:
			generalError = "You don't have permission to perform this action."
		case connect.CodeInvalidArgument:
			// Parse field violations from the error
			fieldErrors = ParseValidationErrors(err)
			if len(fieldErrors) == 0 {
				generalError = "Invalid input provided."
			}
		case connect.CodeNotFound:
			generalError = "The requested resource was not found."
		case connect.CodeResourceExhausted:
			generalError = "Too many requests. Please try again later."
		case connect.CodeInternal:
			generalError = "An internal error occurred. Please try again later."
		default:
			generalError = "An error occurred. Please try again."
		}
	} else if err != nil {
		generalError = "An unexpected error occurred. Please try again."
	}

	return fieldErrors, generalError
}
```

## Handler Implementation Patterns

### 1. Basic Handler Pattern

```go
// UserProfileHandler handles the user profile page
func (h *Handlers) UserProfileHandler(w http.ResponseWriter, r *http.Request) {
	// Get user from session
	session, err := h.sessionStore.Get(r, "user-session")
	if err != nil {
		http.Error(w, "Invalid session", http.StatusBadRequest)
		return
	}

	// Check if user is authenticated
	if session.Values["access_token"] == nil {
		http.Redirect(w, r, "/login", http.StatusSeeOther)
		return
	}

	// Get token and user ID
	token := session.Values["access_token"].(string)
	userID := session.Values["user_id"].(string)

	// Create a context with the token
	ctx := auth.WithToken(r.Context(), token)

	// Create the request
	req := connect.NewRequest(&userv1.GetUserRequest{
		Name: userID,
	})

	// Call the API
	client := h.clients.UserClient
	resp, err := client.GetUser(ctx, req)

	// Handle errors
	if err != nil {
		fieldErrors, generalError := HandleConnectError(err)
		tmplData := map[string]interface{}{
			"Errors":       fieldErrors,
			"GeneralError": generalError,
		}
		h.templates.ExecuteTemplate(w, "error.html", tmplData)
		return
	}

	// Render the template with the user data
	tmplData := map[string]interface{}{
		"User": resp.Msg.User,
	}
	h.templates.ExecuteTemplate(w, "profile.html", tmplData)
}
```

### 2. HTMX Handler Pattern

```go
// CompositionListHandler handles the composition list page/fragment
func (h *Handlers) CompositionListHandler(w http.ResponseWriter, r *http.Request) {
	// Get token from session
	session, _ := h.sessionStore.Get(r, "user-session")
	token := session.Values["access_token"].(string)
	ctx := auth.WithToken(r.Context(), token)

	// Determine if this is an HTMX request
	isHtmx := r.Header.Get("HX-Request") == "true"

	// Parse pagination parameters
	pageSize := 10
	pageToken := r.URL.Query().Get("pageToken")

	// Create the request
	req := connect.NewRequest(&compositionv1.ListCompositionsRequest{
		Parent:    "users/me",
		PageSize:  int32(pageSize),
		PageToken: pageToken,
	})

	// Call the API
	client := h.clients.CompositionClient
	resp, err := client.ListCompositions(ctx, req)

	// Handle errors
	if err != nil {
		if isHtmx {
			// For HTMX requests, return just the error fragment
			fieldErrors, generalError := HandleConnectError(err)
			tmplData := map[string]interface{}{
				"Errors":       fieldErrors,
				"GeneralError": generalError,
			}
			h.templates.ExecuteTemplate(w, "components/error-alert.html", tmplData)
		} else {
			// For full page requests, render the full error page
			fieldErrors, generalError := HandleConnectError(err)
			tmplData := map[string]interface{}{
				"Errors":       fieldErrors,
				"GeneralError": generalError,
			}
			h.templates.ExecuteTemplate(w, "error.html", tmplData)
		}
		return
	}

	// Prepare template data
	tmplData := map[string]interface{}{
		"Compositions": resp.Msg.Compositions,
		"NextPageToken": resp.Msg.NextPageToken,
	}

	// Render the appropriate template based on request type
	if isHtmx {
		// For HTMX requests, return just the list fragment
		h.templates.ExecuteTemplate(w, "components/composition-list.html", tmplData)
	} else {
		// For full page requests, return the full page
		h.templates.ExecuteTemplate(w, "compositions.html", tmplData)
	}
}
```

### 3. Form Submission Pattern

```go
// CreateCompositionHandler handles composition creation
func (h *Handlers) CreateCompositionHandler(w http.ResponseWriter, r *http.Request) {
	// Get token from session
	session, _ := h.sessionStore.Get(r, "user-session")
	token := session.Values["access_token"].(string)
	ctx := auth.WithToken(r.Context(), token)

	// Determine if this is an HTMX request
	isHtmx := r.Header.Get("HX-Request") == "true"

	// Parse form data
	if err := r.ParseForm(); err != nil {
		http.Error(w, "Invalid form data", http.StatusBadRequest)
		return
	}

	// Create the composition request
	req := connect.NewRequest(&compositionv1.CreateCompositionRequest{
		Parent: "users/me",
		Composition: &compositionv1.Composition{
			Title:       r.FormValue("title"),
			Description: r.FormValue("description"),
			Settings: &compositionv1.CompositionSettings{
				PinCount:   parseInt32(r.FormValue("pinCount"), 200),
				LineCount:  parseInt32(r.FormValue("lineCount"), 2000),
				MinWeight:  parseFloat(r.FormValue("minWeight"), 0.1),
				MaxWeight:  parseFloat(r.FormValue("maxWeight"), 1.0),
				Background: r.FormValue("background"),
			},
			ImageId: r.FormValue("imageId"),
		},
	})

	// Call the API
	client := h.clients.CompositionClient
	resp, err := client.CreateComposition(ctx, req)

	// Handle errors
	if err != nil {
		fieldErrors, generalError := HandleConnectError(err)
		tmplData := map[string]interface{}{
			"Errors":       fieldErrors,
			"GeneralError": generalError,
			"FormData":     r.Form, // Include form data for re-rendering
		}

		if isHtmx {
			// For HTMX requests, return just the form with errors
			h.templates.ExecuteTemplate(w, "components/composition-form.html", tmplData)
		} else {
			// For full page requests, render the full page with errors
			h.templates.ExecuteTemplate(w, "create-composition.html", tmplData)
		}
		return
	}

	// On success, redirect or return success fragment
	if isHtmx {
		// For HTMX requests, trigger a redirect via HTMX
		w.Header().Set("HX-Redirect", "/compositions/"+resp.Msg.Composition.Name)
		w.WriteHeader(http.StatusOK)
	} else {
		// For standard requests, use a regular redirect
		http.Redirect(w, r, "/compositions/"+resp.Msg.Composition.Name, http.StatusSeeOther)
	}
}
```

## Advanced Patterns

### 1. Concurrent API Calls

```go
// DashboardHandler fetches multiple resources concurrently
func (h *Handlers) DashboardHandler(w http.ResponseWriter, r *http.Request) {
	// Get token from session
	session, _ := h.sessionStore.Get(r, "user-session")
	token := session.Values["access_token"].(string)
	ctx := auth.WithToken(r.Context(), token)

	// Create channels for results
	type userResult struct {
		user *userv1.User
		err  error
	}
	type compositionsResult struct {
		compositions []*compositionv1.Composition
		err          error
	}
	type statsResult struct {
		stats *statsv1.UserStats
		err   error
	}

	userCh := make(chan userResult, 1)
	compositionsCh := make(chan compositionsResult, 1)
	statsCh := make(chan statsResult, 1)

	// Fetch user profile
	go func() {
		req := connect.NewRequest(&userv1.GetUserRequest{
			Name: "users/me",
		})
		resp, err := h.clients.UserClient.GetUser(ctx, req)
		if err != nil {
			userCh <- userResult{nil, err}
			return
		}
		userCh <- userResult{resp.Msg.User, nil}
	}()

	// Fetch recent compositions
	go func() {
		req := connect.NewRequest(&compositionv1.ListCompositionsRequest{
			Parent:   "users/me",
			PageSize: 5,
		})
		resp, err := h.clients.CompositionClient.ListCompositions(ctx, req)
		if err != nil {
			compositionsCh <- compositionsResult{nil, err}
			return
		}
		compositionsCh <- compositionsResult{resp.Msg.Compositions, nil}
	}()

	// Fetch user stats
	go func() {
		req := connect.NewRequest(&statsv1.GetUserStatsRequest{
			Name: "users/me/stats",
		})
		resp, err := h.clients.StatsClient.GetUserStats(ctx, req)
		if err != nil {
			statsCh <- statsResult{nil, err}
			return
		}
		statsCh <- statsResult{resp.Msg.Stats, nil}
	}()

	// Collect results
	userRes := <-userCh
	compositionsRes := <-compositionsCh
	statsRes := <-statsCh

	// Check for errors
	if userRes.err != nil || compositionsRes.err != nil || statsRes.err != nil {
		var errorMsg string
		if userRes.err != nil {
			_, errorMsg = HandleConnectError(userRes.err)
		} else if compositionsRes.err != nil {
			_, errorMsg = HandleConnectError(compositionsRes.err)
		} else {
			_, errorMsg = HandleConnectError(statsRes.err)
		}

		tmplData := map[string]interface{}{
			"GeneralError": errorMsg,
		}
		h.templates.ExecuteTemplate(w, "error.html", tmplData)
		return
	}

	// Render dashboard with all data
	tmplData := map[string]interface{}{
		"User":         userRes.user,
		"Compositions": compositionsRes.compositions,
		"Stats":        statsRes.stats,
	}
	h.templates.ExecuteTemplate(w, "dashboard.html", tmplData)
}
```

### 2. Streaming Data with Server-Sent Events (SSE)

```go
// CompositionProgressHandler streams progress updates using SSE
func (h *Handlers) CompositionProgressHandler(w http.ResponseWriter, r *http.Request) {
	// Get composition ID from path
	compositionID := chi.URLParam(r, "id")

	// Get token from session
	session, _ := h.sessionStore.Get(r, "user-session")
	token := session.Values["access_token"].(string)
	ctx := auth.WithToken(r.Context(), token)

	// Set up SSE response headers
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.Header().Set("X-Accel-Buffering", "no") // Disable buffering for Nginx

	// Create the streaming request
	req := connect.NewRequest(&compositionv1.WatchCompositionRequest{
		Name: compositionID,
	})

	// Call the streaming API
	client := h.clients.CompositionClient
	stream, err := client.WatchComposition(ctx, req)
	if err != nil {
		_, errMsg := HandleConnectError(err)
		fmt.Fprintf(w, "event: error\ndata: %s\n\n", errMsg)
		return
	}

	// Set up a ticker for heartbeat messages
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	// Create a done channel for cleanup
	done := make(chan bool)
	defer close(done)

	// Handle client disconnection
	go func() {
		<-r.Context().Done()
		done <- true
	}()

	// Stream updates to the client
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "Streaming not supported", http.StatusInternalServerError)
		return
	}

	for {
		select {
		case <-done:
			return
		case <-ticker.C:
			// Send heartbeat to keep connection alive
			fmt.Fprintf(w, ": heartbeat\n\n")
			flusher.Flush()
		default:
			// Try to receive a message
			msg, err := stream.Receive()
			if err != nil {
				if errors.Is(err, io.EOF) {
					// Normal stream end
					fmt.Fprintf(w, "event: complete\ndata: {\"status\":\"completed\"}\n\n")
					flusher.Flush()
					return
				}
				// Error in stream
				_, errMsg := HandleConnectError(err)
				fmt.Fprintf(w, "event: error\ndata: %s\n\n", errMsg)
				flusher.Flush()
				return
			}

			// Send the update
			progress := msg.Msg.Progress
			jsonData, _ := json.Marshal(progress)
			fmt.Fprintf(w, "event: progress\ndata: %s\n\n", jsonData)
			flusher.Flush()

			// If complete, end the stream
			if progress.Status == compositionv1.ProcessingStatus_COMPLETED {
				fmt.Fprintf(w, "event: complete\ndata: {\"status\":\"completed\"}\n\n")
				flusher.Flush()
				return
			}
		}
	}
}
```

## Best Practices

1. **Always use interceptors** for cross-cutting concerns like authentication and logging
2. **Handle errors consistently** using the `HandleConnectError` function
3. **Use context for request-scoped data** like authentication tokens
4. **Implement proper timeout handling** for all API calls
5. **Support both full page and HTMX fragment requests** in handlers
6. **Include form data in error responses** to preserve user input on validation failures
7. **Use concurrent API calls** when appropriate to improve performance
8. **Implement proper session management** with Redis for production environments
9. **Use proper error handling** and context cancellation for streaming endpoints

## Testing

### 1. Unit Testing Connect Clients

```go
func TestUserClient(t *testing.T) {
	// Create a test server that returns mock data
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check the request path
		if r.URL.Path == "/userv1.UserService/GetUser" {
			// Check for auth header
			authHeader := r.Header.Get("Authorization")
			if authHeader != "Bearer test-token" {
				w.WriteHeader(http.StatusUnauthorized)
				return
			}

			// Write a valid Connect response
			w.Header().Set("Content-Type", "application/json")
			w.Write([]byte(`{"msg":{"user":{"name":"users/123","displayName":"Test User","email":"test@example.com"}}}`))
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	})

	server := httptest.NewServer(handler)
	defer server.Close()

	// Create a client config
	config := ClientConfig{
		BaseURL: server.URL,
		Timeout: 5 * time.Second,
	}

	// Create a client
	client := NewUserClient(config)

	// Create a context with auth token
	ctx := auth.WithToken(context.Background(), "test-token")

	// Make a request
	req := connect.NewRequest(&userv1.GetUserRequest{
		Name: "users/123",
	})

	resp, err := client.GetUser(ctx, req)

	// Verify the response
	assert.NoError(t, err)
	assert.NotNil(t, resp)
	assert.Equal(t, "users/123", resp.Msg.User.Name)
	assert.Equal(t, "Test User", resp.Msg.User.DisplayName)
	assert.Equal(t, "test@example.com", resp.Msg.User.Email)
}
```

### 2. Testing Handlers

```go
func TestCompositionHandler(t *testing.T) {
	// Create a mock client
	mockClient := &MockCompositionClient{
		GetCompositionFunc: func(ctx context.Context, req *connect.Request<compositionv1.GetCompositionRequest>) (*connect.Response<compositionv1.GetCompositionResponse>, error) {
			// Check that the request is properly formed
			assert.Equal(t, "compositions/123", req.Msg.Name)

			// Check that auth header is set
			token, ok := auth.TokenFromContext(ctx)
			assert.True(t, ok)
			assert.Equal(t, "test-token", token)

			// Return mock data
			return connect.NewResponse(&compositionv1.GetCompositionResponse{
				Composition: &compositionv1.Composition{
					Name:        "compositions/123",
					Title:       "Test Composition",
					Description: "A test composition",
					Status:      compositionv1.ProcessingStatus_COMPLETED,
				},
			}), nil
		},
	}

	// Create a mock template renderer
	mockTemplates := &MockTemplates{
		ExecuteTemplateFunc: func(w http.ResponseWriter, name string, data interface{}) error {
			// Verify the template name
			assert.Equal(t, "composition.html", name)

			// Verify the template data
			tmplData := data.(map[string]interface{})
			composition := tmplData["Composition"].(*compositionv1.Composition)
			assert.Equal(t, "compositions/123", composition.Name)
			assert.Equal(t, "Test Composition", composition.Title)

			return nil
		},
	}

	// Create a mock session store
	mockSessionStore := &MockSessionStore{
		GetFunc: func(r *http.Request, name string) (*sessions.Session, error) {
			// Return a mock session
			session := &sessions.Session{
				Values: map[interface{}]interface{}{
					"access_token": "test-token",
				},
			}
			return session, nil
		},
	}

	// Create test handler
	handler := &Handlers{
		clients: &Clients{
			CompositionClient: mockClient,
		},
		templates:    mockTemplates,
		sessionStore: mockSessionStore,
	}

	// Create a test request
	req := httptest.NewRequest("GET", "/compositions/123", nil)

	// Create a recorder to capture the response
	w := httptest.NewRecorder()

	// Call the handler
	handler.GetCompositionHandler(w, req)

	// Verify the response
	assert.Equal(t, http.StatusOK, w.Code)
}
```
