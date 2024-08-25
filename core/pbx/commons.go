package pbx

import (
	"errors"
	"strings"

	"github.com/google/uuid"
)

type ResourceType struct {
	Parent *ResourceType
	Type   string
}

type Resource struct {
	Type *ResourceType
	ID   string
}

const (
	RessourceNameUsers = "users"
	RessourceNameArts  = "arts"
)

var validResourceTypesList = []string{
	RessourceNameUsers,
	RessourceNameArts,
}

var (
	RessourceTypeUsers = &ResourceType{Type: RessourceNameUsers}
	RessourceTypeArts  = &ResourceType{Type: RessourceNameArts, Parent: RessourceTypeUsers}
)

// GetResourcesFromResourceName splits the given resourceName string by "/"
// and constructs a map of ResourceType to ID. Each pair of consecutive parts
// in the resourceName represents a ResourceType and its corresponding ID.
// The function returns the constructed map of resources and an error if the
// resourceName is invalid (i.e., the number of parts is not even).
func GetResourcesFromResourceName(resourceName string) (map[string]string, error) {
	parts := strings.Split(resourceName, "/")
	if len(parts)%2 != 0 {
		return nil, errors.New("invalid resource name")
	}

	resources := make(map[string]string)
	for i := 0; i < len(parts); i += 2 {
		//check if the ID is a valid UUID
		if _, err := uuid.Parse(parts[i+1]); err != nil {
			return nil, errors.New("invalid resource ID for resource type " + parts[i])
		}
		if !isResourceTypeValid(parts[i]) {
			return nil, errors.New("invalid resource type " + parts[i])
		}
		resources[parts[i]] = parts[i+1]
	}

	return resources, nil
}

// GetResourceIDByType returns the ID of a resource with the given name and type.
// It searches for the resource in the list of resources obtained from the resource name.
// If the resource is found, its ID is returned. Otherwise, an error is returned.
func GetResourceIDByType(resourceName string, resourceType *ResourceType) (string, error) {
	resources, err := GetResourcesFromResourceName(resourceName)
	if err != nil {
		return "", err
	}

	for ressourceType, resourceId := range resources {
		if ressourceType == resourceType.Type {
			return resourceId, nil
		}
	}

	return "", errors.New("resource not found")
}

// GetResourceName concatenates the type and ID of each resource in the given slice
// and returns the resulting string. The resources are expected to have a "Type" field
// and an "ID" field.
func GetResourceName(resources []Resource) string {
	var builder strings.Builder

	for _, resource := range resources {
		builder.WriteString(resource.Type.Type)
		builder.WriteString("/")
		builder.WriteString(resource.ID)
		builder.WriteString("/")
	}

	result := builder.String()
	if len(result) > 0 {
		result = result[:len(result)-1]
	}

	return result
}

func isResourceTypeValid(resourceType string) bool {
	for _, validResourceType := range validResourceTypesList {
		if resourceType == validResourceType {
			return true
		}
	}
	return false
}
