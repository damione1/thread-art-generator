package templates

import (
	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/Damione1/thread-art-generator/client/internal/types"
	"github.com/Damione1/thread-art-generator/core/pb"
)

// PageData represents the centralized data structure for all templates
// This follows the View Model pattern recommended in Templ documentation
type PageData struct {
	// Core page information
	Title    string
	PageType string
	
	// User information (can be nil for public pages)
	User *auth.UserInfo
	
	// Authentication state
	IsLoggedIn bool
	
	// Optional Firebase configuration for auth pages
	FirebaseConfig *types.FirebaseConfig
	
	// Meta information and SEO
	Meta map[string]string
	
	// Page-specific data (can hold any data specific to the page)
	Data interface{}
	
	// HTMX-related data
	HTMXRequest bool
	HTMXTarget  string
	
	// Error handling
	ErrorMessage string
	FieldErrors  map[string][]string
}

// NewPageData creates a new PageData with sensible defaults
func NewPageData(title, pageType string) *PageData {
	return &PageData{
		Title:       title,
		PageType:    pageType,
		User:        nil,
		IsLoggedIn:  false,
		Meta:        make(map[string]string),
		FieldErrors: make(map[string][]string),
	}
}

// WithUser adds user information to the page data
func (pd *PageData) WithUser(user *auth.UserInfo) *PageData {
	pd.User = user
	pd.IsLoggedIn = user != nil
	return pd
}

// WithFirebaseConfig adds Firebase configuration to the page data
func (pd *PageData) WithFirebaseConfig(config *types.FirebaseConfig) *PageData {
	pd.FirebaseConfig = config
	return pd
}

// WithData adds page-specific data
func (pd *PageData) WithData(data interface{}) *PageData {
	pd.Data = data
	return pd
}

// WithError adds an error message to the page data
func (pd *PageData) WithError(message string) *PageData {
	pd.ErrorMessage = message
	return pd
}

// WithFieldErrors adds field-specific errors to the page data
func (pd *PageData) WithFieldErrors(errors map[string][]string) *PageData {
	pd.FieldErrors = errors
	return pd
}

// WithMeta adds meta information
func (pd *PageData) WithMeta(key, value string) *PageData {
	pd.Meta[key] = value
	return pd
}

// WithHTMX adds HTMX-related information
func (pd *PageData) WithHTMX(target string) *PageData {
	pd.HTMXRequest = true
	pd.HTMXTarget = target
	return pd
}

// Helper methods for templates

// HasError returns true if there's a general error message
func (pd *PageData) HasError() bool {
	return pd.ErrorMessage != ""
}

// HasFieldErrors returns true if there are field-specific errors
func (pd *PageData) HasFieldErrors() bool {
	return len(pd.FieldErrors) > 0
}

// GetFieldErrors returns errors for a specific field
func (pd *PageData) GetFieldErrors(field string) []string {
	return pd.FieldErrors[field]
}

// HasFieldError returns true if a specific field has errors
func (pd *PageData) HasFieldError(field string) bool {
	errors, exists := pd.FieldErrors[field]
	return exists && len(errors) > 0
}

// GetPageTitle returns the full page title with fallback
func (pd *PageData) GetPageTitle() string {
	if pd.Title != "" {
		return pd.Title
	}
	return "ThreadArt - Create Beautiful Thread Art"
}

// Safe user data access helpers

// SafeUserDisplayName returns the display name with fallback hierarchy
func SafeUserDisplayName(user *auth.UserInfo) string {
	if user == nil {
		return "Guest"
	}
	if user.Name != "" {
		return user.Name
	}
	if user.Email != "" {
		return user.Email
	}
	return "Unknown User"
}

// SafeUserInitials returns the first character of user name, email, or default placeholder
func SafeUserInitials(user *auth.UserInfo) string {
	if user == nil {
		return "?"
	}
	if user.Name != "" && len([]rune(user.Name)) > 0 {
		return string([]rune(user.Name)[0])
	}
	if user.Email != "" && len([]rune(user.Email)) > 0 {
		return string([]rune(user.Email)[0])
	}
	return "U"
}

// GetUserDisplayName safely returns the user's display name with fallbacks
func (pd *PageData) GetUserDisplayName() string {
	return SafeUserDisplayName(pd.User)
}

// GetUserInitials safely returns the user's initials with fallbacks
func (pd *PageData) GetUserInitials() string {
	return SafeUserInitials(pd.User)
}

// GetDashboardData safely returns the dashboard data from the page data
func (pd *PageData) GetDashboardData() *DashboardPageData {
	if pd.Data == nil {
		return &DashboardPageData{}
	}
	if dashData, ok := pd.Data.(*DashboardPageData); ok {
		return dashData
	}
	return &DashboardPageData{}
}

// Specific page data types

// DashboardPageData contains data specific to the dashboard page
type DashboardPageData struct {
	Arts interface{} // Will be []*pb.Art from pb
	Sort string
	Dir  string
}

// GetArts safely returns the arts as the expected type
func (d *DashboardPageData) GetArts() []*pb.Art {
	if d.Arts == nil {
		return []*pb.Art{}
	}
	if arts, ok := d.Arts.([]*pb.Art); ok {
		return arts
	}
	return []*pb.Art{}
}

// ArtPageData contains data specific to art-related pages
type ArtPageData struct {
	Art        interface{} // Will be the actual art type from pb
	UploadURL  string
	IsEditing  bool
}

// AuthPageData contains data specific to authentication pages
type AuthPageData struct {
	ReturnURL      string
	HasError       bool
	ErrorMessage   string
	FieldErrors    map[string][]string
}