package templates

import (
	"testing"

	"github.com/Damione1/thread-art-generator/client/internal/auth"
	"github.com/stretchr/testify/require"
)

func TestSafeUserDisplayName(t *testing.T) {
	require.Equal(t, "Guest", SafeUserDisplayName(nil))
	require.Equal(t, "Ada Lovelace", SafeUserDisplayName(&auth.UserInfo{
		FirstName: "Ada",
		LastName:  "Lovelace",
		Email:     "ada@example.com",
		Name:      "ada@example.com",
	}))
	require.Equal(t, "ada@example.com", SafeUserDisplayName(&auth.UserInfo{
		Email: "ada@example.com",
		Name:  "ada@example.com",
	}))
	require.Equal(t, "Unknown User", SafeUserDisplayName(&auth.UserInfo{}))
}

func TestSafeUserInitials(t *testing.T) {
	require.Equal(t, "?", SafeUserInitials(nil))
	require.Equal(t, "A", SafeUserInitials(&auth.UserInfo{FirstName: "Ada", Email: "ada@example.com"}))
	require.Equal(t, "A", SafeUserInitials(&auth.UserInfo{Email: "ada@example.com"}))
}
