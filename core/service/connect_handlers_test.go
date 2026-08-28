package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/stretchr/testify/require"
)

func TestServerUnimplementedUserAdmin(t *testing.T) {
	t.Parallel()
	s := &Server{}

	_, err := s.ListUsers(context.Background(), connect.NewRequest(&pb.ListUsersRequest{}))
	require.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))

	_, err = s.DeleteUser(context.Background(), connect.NewRequest(&pb.DeleteUserRequest{}))
	require.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
}
