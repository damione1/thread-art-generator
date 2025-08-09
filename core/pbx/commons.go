package pbx

import (
	"strings"
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
	RessourceNameUsers        = "users"
	RessourceNameArts         = "arts"
	RessourceNameCompositions = "compositions"
)

var validResourceTypesList = []string{
	RessourceNameUsers,
	RessourceNameArts,
	RessourceNameCompositions,
}

var (
	RessourceTypeUsers        = &ResourceType{Type: RessourceNameUsers}
	RessourceTypeArts         = &ResourceType{Type: RessourceNameArts, Parent: RessourceTypeUsers}
	RessourceTypeCompositions = &ResourceType{Type: RessourceNameCompositions, Parent: RessourceTypeArts}
)

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
