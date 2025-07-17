package resource

import (
	"fmt"

	"go.einride.tech/aip/resourcename"
)

const (
	UserResource        string = "users/{user}"
	ArtResource         string = "users/{user}/arts/{art}"
	CompositionResource string = "users/{user}/arts/{art}/compositions/{composition}"
)

// ResourceParser interface for parsing resource names
type ResourceParser interface {
	Parse(resourceName string) (Resource, error)
	ResourceType(resourceName string) (string, error)
}

// Parser implements ResourceParser interface
type Parser struct{}

// NewParser creates a new resource parser
func NewParser() ResourceParser {
	return &Parser{}
}

type Resource interface{}

type User struct {
	ID string
}

type Art struct {
	UserID string
	ArtID  string
}

type Composition struct {
	UserID        string
	ArtID         string
	CompositionID string
}

// Builder functions for creating resource names
func BuildUserResourceName(userID string) string {
	return fmt.Sprintf("users/%s", userID)
}

func BuildArtResourceName(userID, artID string) string {
	return fmt.Sprintf("users/%s/arts/%s", userID, artID)
}

func BuildCompositionResourceName(userID, artID, compositionID string) string {
	return fmt.Sprintf("users/%s/arts/%s/compositions/%s", userID, artID, compositionID)
}

// Parse parses a resource name and returns the appropriate resource type
func (p *Parser) Parse(resourceName string) (Resource, error) {
	if err := validateResourceName(resourceName); err != nil {
		return nil, fmt.Errorf("invalid resource name: %v", err)
	}

	pattern, err := p.ResourceType(resourceName)
	if err != nil {
		return nil, err
	}

	switch pattern {
	case UserResource:
		return p.parseUserResource(resourceName)
	case ArtResource:
		return p.parseArtResource(resourceName)
	case CompositionResource:
		return p.parseCompositionResource(resourceName)
	default:
		return nil, fmt.Errorf("invalid resource type")
	}
}

// ResourceType determines the resource type pattern for a given name
func (p *Parser) ResourceType(resourceName string) (string, error) {
	switch {
	case resourcename.Match(UserResource, resourceName):
		return UserResource, nil
	case resourcename.Match(ArtResource, resourceName):
		return ArtResource, nil
	case resourcename.Match(CompositionResource, resourceName):
		return CompositionResource, nil
	default:
		return "", fmt.Errorf("invalid resource name")
	}
}

func (p *Parser) parseUserResource(resourceName string) (*User, error) {
	var userID string
	err := resourcename.Sscan(resourceName, UserResource, &userID)
	if err != nil {
		return nil, err
	}

	return &User{ID: userID}, nil
}

func (p *Parser) parseArtResource(resourceName string) (*Art, error) {
	var userID, artID string
	err := resourcename.Sscan(resourceName, ArtResource, &userID, &artID)
	if err != nil {
		return nil, err
	}

	return &Art{
		UserID: userID,
		ArtID:  artID,
	}, nil
}

func (p *Parser) parseCompositionResource(resourceName string) (*Composition, error) {
	var userID, artID, compositionID string
	err := resourcename.Sscan(resourceName, CompositionResource, &userID, &artID, &compositionID)
	if err != nil {
		return nil, err
	}

	return &Composition{
		UserID:        userID,
		ArtID:         artID,
		CompositionID: compositionID,
	}, nil
}

func validateResourceName(name string) error {
	return resourcename.Validate(name)
}

// Convenience functions that use the default parser
var defaultParser = NewParser()

// ParseResourceName is a convenience function using the default parser
func ParseResourceName(resourceName string) (Resource, error) {
	return defaultParser.Parse(resourceName)
}

// GetResourceType is a convenience function using the default parser
func GetResourceType(resourceName string) (string, error) {
	return defaultParser.ResourceType(resourceName)
}

// ExtractUserID extracts the user ID from any resource name that contains a user
func ExtractUserID(resourceName string) string {
	// Try to parse as user resource first
	if resource, err := defaultParser.Parse(resourceName); err == nil {
		switch r := resource.(type) {
		case *User:
			return r.ID
		case *Art:
			return r.UserID
		case *Composition:
			return r.UserID
		}
	}
	
	// Fallback: extract using resourcename scanning
	var userID string
	if err := resourcename.Sscan(resourceName, UserResource, &userID); err == nil {
		return userID
	}
	
	var artID string
	if err := resourcename.Sscan(resourceName, ArtResource, &userID, &artID); err == nil {
		return userID
	}
	
	var compositionID string
	if err := resourcename.Sscan(resourceName, CompositionResource, &userID, &artID, &compositionID); err == nil {
		return userID
	}
	
	// Return as-is if not a valid pattern (for backward compatibility)
	return resourceName
}
