package auth

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/auth0/go-jwt-middleware/v2/jwks"
	"github.com/auth0/go-jwt-middleware/v2/validator"
	"github.com/rs/zerolog/log"
)

// Auth0Configuration holds Auth0-specific configuration
type Auth0Configuration struct {
	Domain                    string
	Audience                  string
	ClientID                  string
	ClientSecret              string
	ManagementApiClientID     string
	ManagementApiClientSecret string
}

// Auth0Service implements both Authenticator and UserProvider interfaces
type Auth0Service struct {
	config     Auth0Configuration
	validator  *validator.Validator
	httpClient *http.Client
}

// customClaims contains custom data we want from the Auth0 token
type customClaims struct {
	Auth0ID string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}

// Validate does nothing but is required for the validator interface
func (c customClaims) Validate(ctx context.Context) error {
	return nil
}

// NewAuth0Service creates a new Auth0 service implementing AuthService
func NewAuth0Service(config Auth0Configuration) (AuthService, error) {
	issuerURL, err := url.Parse(fmt.Sprintf("https://%s/", config.Domain))
	if err != nil {
		return nil, fmt.Errorf("failed to parse issuer URL: %v", err)
	}

	provider := jwks.NewCachingProvider(issuerURL, 5*time.Minute)

	jwtValidator, err := validator.New(
		provider.KeyFunc,
		validator.RS256,
		issuerURL.String(),
		[]string{config.Audience},
		validator.WithCustomClaims(
			func() validator.CustomClaims {
				return &customClaims{}
			},
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create validator: %v", err)
	}

	return &Auth0Service{
		config:     config,
		validator:  jwtValidator,
		httpClient: &http.Client{Timeout: 10 * time.Second},
	}, nil
}

// ValidateToken validates the token and returns the claims
func (a *Auth0Service) ValidateToken(ctx context.Context, tokenString string) (*AuthClaims, error) {
	claims, err := a.validator.ValidateToken(ctx, tokenString)
	if err != nil {
		return nil, err
	}

	validatedClaims, ok := claims.(*validator.ValidatedClaims)
	if !ok {
		return nil, fmt.Errorf("failed to cast claims to ValidatedClaims")
	}

	customClaims, ok := validatedClaims.CustomClaims.(*customClaims)
	if !ok {
		return nil, fmt.Errorf("failed to cast to custom claims")
	}

	// Create basic claims from the token
	authClaims := &AuthClaims{
		UserID:  customClaims.Auth0ID,
		Email:   customClaims.Email,
		Name:    customClaims.Name,
		Picture: customClaims.Picture,
	}

	// If email or name is missing, fetch from userinfo endpoint
	if authClaims.Email == "" || authClaims.Name == "" {
		userInfo, err := a.fetchUserInfo(ctx, tokenString)
		if err != nil {
			log.Warn().Err(err).Str("user_id", authClaims.UserID).Msg("Failed to fetch user info, continuing with limited claims")
		} else {
			// Update with data from userinfo endpoint
			if authClaims.Email == "" {
				authClaims.Email = userInfo.Email
			}
			if authClaims.Name == "" {
				authClaims.Name = userInfo.Name
			}
			if authClaims.Picture == "" {
				authClaims.Picture = userInfo.Picture
			}
		}
	}

	return authClaims, nil
}

// fetchUserInfo fetches user profile data from Auth0 userinfo endpoint
func (a *Auth0Service) fetchUserInfo(ctx context.Context, tokenString string) (*struct {
	Sub     string `json:"sub"`
	Email   string `json:"email"`
	Name    string `json:"name"`
	Picture string `json:"picture"`
}, error) {
	// Create request to Auth0 userinfo endpoint
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("https://%s/userinfo", a.config.Domain),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create userinfo request: %v", err)
	}

	// Add Authorization header
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", tokenString))

	// Execute request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch userinfo: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("userinfo returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var userInfo struct {
		Sub     string `json:"sub"`
		Email   string `json:"email"`
		Name    string `json:"name"`
		Picture string `json:"picture"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode userinfo response: %v", err)
	}

	return &userInfo, nil
}

// GetAuthMiddleware returns the validator for middleware integration
func (a *Auth0Service) GetAuthMiddleware() interface{} {
	return a.validator
}

// GetUserInfoFromToken retrieves user information directly from Auth0 userinfo endpoint using a token
// This is a simpler alternative to the Management API when you already have a user token
func (a *Auth0Service) GetUserInfoFromToken(ctx context.Context, token string) (*UserInfo, error) {
	// Extract the auth0 identifier from token claims
	claims, err := a.ValidateToken(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to validate token: %v", err)
	}

	// Extract provider from Auth0 ID
	parts := strings.Split(claims.UserID, "|")
	provider := "auth0"
	if len(parts) > 1 {
		provider = parts[0]
	}

	// Try userinfo endpoint
	userInfo, err := a.fetchUserInfo(ctx, token)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user info: %v", err)
	}

	// Special case for GitHub users - they often don't share email by default
	if provider == "github" && userInfo.Email == "" {
		// Try to fetch email from GitHub API directly if we have Management API access
		if a.config.ManagementApiClientID != "" && a.config.ManagementApiClientSecret != "" {
			githubUserID := ""
			if len(parts) > 1 {
				githubUserID = parts[1]
				log.Debug().Str("github_id", githubUserID).Msg("Trying to fetch GitHub email using Management API")

				// First get a management token
				mgmtToken, err := a.getManagementToken(ctx)
				if err == nil {
					// Use the Management API to get the user's identities
					mReq, err := http.NewRequestWithContext(
						ctx,
						"GET",
						fmt.Sprintf("https://%s/api/v2/users/%s", a.config.Domain, url.PathEscape(claims.UserID)),
						nil,
					)
					if err == nil {
						mReq.Header.Add("Authorization", fmt.Sprintf("Bearer %s", mgmtToken))

						mResp, err := a.httpClient.Do(mReq)
						if err == nil {
							defer mResp.Body.Close()

							if mResp.StatusCode == http.StatusOK {
								var userObj struct {
									Email      string `json:"email"`
									GithubMeta struct {
										Email       string `json:"email"`
										PublicEmail string `json:"public_email"`
									} `json:"github_meta"`
									Identities []struct {
										Provider    string `json:"provider"`
										UserID      string `json:"user_id"`
										IsSocial    bool   `json:"isSocial"`
										ProfileData struct {
											Email string `json:"email"`
										} `json:"profileData"`
									} `json:"identities"`
								}

								if err := json.NewDecoder(mResp.Body).Decode(&userObj); err == nil {
									// Check if we got an email from the user object directly
									if userObj.Email != "" {
										userInfo.Email = userObj.Email
										log.Debug().Str("email", userObj.Email).Msg("Found email in user object")
									} else {
										// Check identities
										for _, identity := range userObj.Identities {
											if identity.Provider == "github" && identity.ProfileData.Email != "" {
												userInfo.Email = identity.ProfileData.Email
												log.Debug().Str("email", identity.ProfileData.Email).Msg("Found email in identity profile data")
												break
											}
										}

										// If we still don't have an email, check the GitHub metadata
										if userInfo.Email == "" && userObj.GithubMeta.Email != "" {
											userInfo.Email = userObj.GithubMeta.Email
											log.Debug().Str("email", userObj.GithubMeta.Email).Msg("Found email in GitHub metadata")
										}

										if userInfo.Email == "" && userObj.GithubMeta.PublicEmail != "" {
											userInfo.Email = userObj.GithubMeta.PublicEmail
											log.Debug().Str("email", userObj.GithubMeta.PublicEmail).Msg("Found public email in GitHub metadata")
										}
									}
								}
							}
						}
					}
				}
			}
		}

		// If we still don't have an email, use a placeholder
		if userInfo.Email == "" {
			log.Warn().Str("user_id", claims.UserID).Msg("Unable to retrieve email for GitHub user, using placeholder")
			userInfo.Email = fmt.Sprintf("github-%s@placeholder.local", parts[1])
		}
	}

	// Parse name into first/last name components if possible
	firstName, lastName := "", ""
	if userInfo.Name != "" {
		nameParts := strings.SplitN(userInfo.Name, " ", 2)
		if len(nameParts) > 0 {
			firstName = nameParts[0]
			if len(nameParts) > 1 {
				lastName = nameParts[1]
			}
		}
	}

	return &UserInfo{
		ID:        claims.UserID,
		Email:     userInfo.Email,
		Name:      userInfo.Name,
		FirstName: firstName,
		LastName:  lastName,
		Picture:   userInfo.Picture,
		CreatedAt: time.Now().Format(time.RFC3339), // Userinfo doesn't provide timestamps
		UpdatedAt: time.Now().Format(time.RFC3339),
		Provider:  provider,
	}, nil
}

// GetUserInfo retrieves user information from Auth0
func (a *Auth0Service) GetUserInfo(ctx context.Context, userID string) (*UserInfo, error) {
	// Extract the auth0 identifier - typically in format "provider|userid"
	parts := strings.Split(userID, "|")
	provider := "auth0"
	if len(parts) > 1 {
		provider = parts[0]
	}

	// Try to get the user info from Auth0 Management API
	userInfo, err := a.fetchUserInfoFromManagement(ctx, userID)
	if err != nil {
		log.Warn().Err(err).Str("user_id", userID).Msg("Failed to fetch user from Management API, trying context token")

		// Try to extract token from context
		if token, ok := ctx.Value("token").(string); ok && token != "" {
			// Try userinfo endpoint as fallback
			log.Debug().Str("user_id", userID).Msg("Found token in context, trying userinfo endpoint")

			uiInfo, uiErr := a.fetchUserInfo(ctx, token)
			if uiErr != nil {
				log.Warn().Err(uiErr).Str("user_id", userID).Msg("Failed to fetch user from userinfo endpoint")
			} else {
				// Parse name into first/last name components if possible
				firstName, lastName := "", ""
				if uiInfo.Name != "" {
					nameParts := strings.SplitN(uiInfo.Name, " ", 2)
					if len(nameParts) > 0 {
						firstName = nameParts[0]
						if len(nameParts) > 1 {
							lastName = nameParts[1]
						}
					}
				}

				return &UserInfo{
					ID:        userID,
					Email:     uiInfo.Email,
					Name:      uiInfo.Name,
					FirstName: firstName,
					LastName:  lastName,
					Picture:   uiInfo.Picture,
					CreatedAt: time.Now().Format(time.RFC3339), // Userinfo doesn't provide timestamps
					UpdatedAt: time.Now().Format(time.RFC3339),
					Provider:  provider,
				}, nil
			}
		}

		// Fall back to minimal info if both methods fail
		log.Warn().Str("user_id", userID).Msg("Using minimal profile as fallback")
		return &UserInfo{
			ID:        userID,
			Email:     "",
			Name:      "",
			FirstName: "",
			LastName:  "",
			Picture:   "",
			CreatedAt: time.Now().Format(time.RFC3339),
			UpdatedAt: time.Now().Format(time.RFC3339),
			Provider:  provider,
		}, nil
	}

	// Parse name into first/last name components if possible
	firstName, lastName := "", ""
	if userInfo.Name != "" {
		nameParts := strings.SplitN(userInfo.Name, " ", 2)
		if len(nameParts) > 0 {
			firstName = nameParts[0]
			if len(nameParts) > 1 {
				lastName = nameParts[1]
			}
		}
	}

	return &UserInfo{
		ID:        userID,
		Email:     userInfo.Email,
		Name:      userInfo.Name,
		FirstName: firstName,
		LastName:  lastName,
		Picture:   userInfo.Picture,
		CreatedAt: userInfo.CreatedAt,
		UpdatedAt: userInfo.UpdatedAt,
		Provider:  provider,
	}, nil
}

// fetchUserInfoFromManagement fetches user info from Auth0 Management API
func (a *Auth0Service) fetchUserInfoFromManagement(ctx context.Context, userID string) (*struct {
	Email     string `json:"email"`
	Name      string `json:"name"`
	Picture   string `json:"picture"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}, error) {
	// First, get a management API token
	token, err := a.getManagementToken(ctx)
	if err != nil {
		return nil, err
	}

	// Create request to Auth0 Management API
	req, err := http.NewRequestWithContext(
		ctx,
		"GET",
		fmt.Sprintf("https://%s/api/v2/users/%s", a.config.Domain, url.PathEscape(userID)),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create user request: %v", err)
	}

	// Add Authorization header with management token
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", token))

	// Execute request
	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch user: %v", err)
	}
	defer resp.Body.Close()

	// Check response status
	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("user request returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse the response
	var userInfo struct {
		Email     string `json:"email"`
		Name      string `json:"name"`
		Picture   string `json:"picture"`
		CreatedAt string `json:"created_at"`
		UpdatedAt string `json:"updated_at"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&userInfo); err != nil {
		return nil, fmt.Errorf("failed to decode user response: %v", err)
	}

	return &userInfo, nil
}

// getManagementToken gets an access token for the Auth0 Management API
func (a *Auth0Service) getManagementToken(ctx context.Context) (string, error) {
	// Debug log the management API configuration but mask secrets
	log.Debug().
		Str("domain", a.config.Domain).
		Str("management_client_id", a.config.ManagementApiClientID).
		Str("management_client_secret_length", fmt.Sprintf("%d chars", len(a.config.ManagementApiClientSecret))).
		Msg("Getting management API token")

	// Check if management credentials are set
	if a.config.ManagementApiClientID == "" || a.config.ManagementApiClientSecret == "" {
		return "", fmt.Errorf("management API credentials not set")
	}

	data := url.Values{}
	data.Set("grant_type", "client_credentials")
	data.Set("client_id", a.config.ManagementApiClientID)
	data.Set("client_secret", a.config.ManagementApiClientSecret)

	// IMPORTANT: The audience for the Management API must be fixed format
	data.Set("audience", fmt.Sprintf("https://%s/api/v2/", a.config.Domain))

	// Debug request data
	reqBody := data.Encode()
	log.Debug().
		Str("endpoint", fmt.Sprintf("https://%s/oauth/token", a.config.Domain)).
		Str("audience", fmt.Sprintf("https://%s/api/v2/", a.config.Domain)).
		Msg("Auth0 management token request")

	req, err := http.NewRequestWithContext(
		ctx,
		"POST",
		fmt.Sprintf("https://%s/oauth/token", a.config.Domain),
		strings.NewReader(reqBody),
	)
	if err != nil {
		return "", fmt.Errorf("failed to create token request: %v", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to get management token: %v", err)
	}
	defer resp.Body.Close()

	bodyBytes, _ := io.ReadAll(resp.Body)

	if resp.StatusCode != http.StatusOK {
		// Log both status code and response body for debugging
		log.Debug().
			Int("status_code", resp.StatusCode).
			Str("response", string(bodyBytes)).
			Msg("Auth0 management token request failed")

		return "", fmt.Errorf("token request returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Re-create reader from bodyBytes since we've consumed it
	var tokenResp struct {
		AccessToken string `json:"access_token"`
		TokenType   string `json:"token_type"`
		ExpiresIn   int    `json:"expires_in"`
	}

	if err := json.Unmarshal(bodyBytes, &tokenResp); err != nil {
		return "", fmt.Errorf("failed to decode token response: %v", err)
	}

	log.Debug().
		Str("token_type", tokenResp.TokenType).
		Int("expires_in", tokenResp.ExpiresIn).
		Msg("Auth0 management token obtained successfully")

	return tokenResp.AccessToken, nil
}

func (a *Auth0Service) UpdateUserPassword(ctx context.Context, userID string, newPassword string) error {
	// With SPA authentication, password updates should be handled client-side
	// This method becomes a no-op in this implementation
	log.Warn().Str("user_id", userID).Msg("UpdateUserPassword called but not implemented in SPA mode")
	return nil
}

// UpdateUserProfile updates a user's profile information in Auth0
func (a *Auth0Service) UpdateUserProfile(ctx context.Context, userID string, profile UserProfile) error {
	// With SPA authentication, profile updates should be handled client-side
	// This method becomes a no-op in this implementation
	log.Warn().
		Str("user_id", userID).
		Str("name", profile.Name).
		Msg("UpdateUserProfile called but not implemented in SPA mode")
	return nil
}
