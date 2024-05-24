package resource

import (
	"fmt"

	"go.einride.tech/aip/resourcename"
)

const (
	UserResource    = "users/{user}"
	ArtResource     = "users/{user}/arts/{art}"
	AttemptResource = "users/{user}/arts/{art}/attempts/{attempt}"
)

type ResourceParser struct {
	name    string
	pattern string
}

type Resource interface{}

type User struct {
	ID string
}

type Art struct {
	UserID string
	ArtID  string
}

type Attempt struct {
	UserID    string
	ArtID     string
	AttemptID string
}

func NewResourceParser(name string) (*ResourceParser, error) {
	if err := validateResourceName(name); err != nil {
		return nil, fmt.Errorf("invalid resource name: %v", err)
	}

	pattern, err := getResourceType(name)
	if err != nil {
		return nil, err
	}

	return &ResourceParser{
		name:    name,
		pattern: pattern,
	}, nil
}

func (rp *ResourceParser) Parse() (Resource, error) {
	switch rp.pattern {
	case UserResource:
		return rp.parseUserResource()
	case ArtResource:
		return rp.parseArtResource()
	case AttemptResource:
		return rp.parseAttemptResource()
	default:
		return nil, fmt.Errorf("invalid resource type")
	}
}

func (rp *ResourceParser) parseUserResource() (*User, error) {
	var userID string
	err := resourcename.Sscan(rp.name, UserResource, &userID)
	if err != nil {
		return nil, err
	}

	return &User{ID: userID}, nil
}

func (rp *ResourceParser) parseArtResource() (*Art, error) {
	var userID, artID string
	err := resourcename.Sscan(rp.name, ArtResource, &userID, &artID)
	if err != nil {
		return nil, err
	}

	return &Art{
		UserID: userID,
		ArtID:  artID,
	}, nil
}

func (rp *ResourceParser) parseAttemptResource() (*Attempt, error) {
	var userID, artID, attemptID string
	err := resourcename.Sscan(rp.name, AttemptResource, &userID, &artID, &attemptID)
	if err != nil {
		return nil, err
	}

	return &Attempt{
		UserID:    userID,
		ArtID:     artID,
		AttemptID: attemptID,
	}, nil
}

func (rp *ResourceParser) ResourceType() string {
	return rp.pattern
}

func validateResourceName(name string) error {
	return resourcename.Validate(name)
}

func getResourceType(name string) (string, error) {
	switch {
	case resourcename.Match(UserResource, name):
		return UserResource, nil
	case resourcename.Match(ArtResource, name):
		return ArtResource, nil
	case resourcename.Match(AttemptResource, name):
		return AttemptResource, nil
	default:
		return "", fmt.Errorf("invalid resource name")
	}
}
