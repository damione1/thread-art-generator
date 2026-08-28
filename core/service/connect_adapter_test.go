package service

import (
	"context"
	"testing"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/stretchr/testify/require"
)

func TestConnectAdapterUnimplementedUserAdmin(t *testing.T) {
	t.Parallel()
	a := NewConnectAdapter(nil)

	_, err := a.ListUsers(context.Background(), connect.NewRequest(&pb.ListUsersRequest{}))
	require.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))

	_, err = a.DeleteUser(context.Background(), connect.NewRequest(&pb.DeleteUserRequest{}))
	require.Equal(t, connect.CodeUnimplemented, connect.CodeOf(err))
}
