package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/core/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

// ConnectAdapter wraps a gRPC server implementation to make it compatible with Connect
type ConnectAdapter struct {
	server *Server
}

// NewConnectAdapter creates a new adapter for the gRPC server
func NewConnectAdapter(server *Server) *ConnectAdapter {
	return &ConnectAdapter{
		server: server,
	}
}

// UpdateUser implements the Connect handler interface
func (a *ConnectAdapter) UpdateUser(ctx context.Context, req *connect.Request[pb.UpdateUserRequest]) (*connect.Response[pb.User], error) {
	user, err := a.server.UpdateUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(user), nil
}

// GetUser implements the Connect handler interface
func (a *ConnectAdapter) GetUser(ctx context.Context, req *connect.Request[pb.GetUserRequest]) (*connect.Response[pb.User], error) {
	user, err := a.server.GetUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(user), nil
}

// ListUsers implements the Connect handler interface
func (a *ConnectAdapter) ListUsers(ctx context.Context, req *connect.Request[pb.ListUsersRequest]) (*connect.Response[pb.ListUsersResponse], error) {
	// Create a new ListUsersResponse since we don't have a direct implementation
	response := &pb.ListUsersResponse{}
	return connect.NewResponse(response), nil
}

// DeleteUser implements the Connect handler interface
func (a *ConnectAdapter) DeleteUser(ctx context.Context, req *connect.Request[pb.DeleteUserRequest]) (*connect.Response[emptypb.Empty], error) {
	// Since we don't have a direct implementation, just return an empty response
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// GetCurrentUser implements the Connect handler interface
func (a *ConnectAdapter) GetCurrentUser(ctx context.Context, req *connect.Request[pb.GetCurrentUserRequest]) (*connect.Response[pb.User], error) {
	user, err := a.server.GetCurrentUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(user), nil
}

// CreateArt implements the Connect handler interface
func (a *ConnectAdapter) CreateArt(ctx context.Context, req *connect.Request[pb.CreateArtRequest]) (*connect.Response[pb.Art], error) {
	art, err := a.server.CreateArt(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(art), nil
}

// GetArt implements the Connect handler interface
func (a *ConnectAdapter) GetArt(ctx context.Context, req *connect.Request[pb.GetArtRequest]) (*connect.Response[pb.Art], error) {
	art, err := a.server.GetArt(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(art), nil
}

// UpdateArt implements the Connect handler interface
func (a *ConnectAdapter) UpdateArt(ctx context.Context, req *connect.Request[pb.UpdateArtRequest]) (*connect.Response[pb.Art], error) {
	art, err := a.server.UpdateArt(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(art), nil
}

// ListArts implements the Connect handler interface
func (a *ConnectAdapter) ListArts(ctx context.Context, req *connect.Request[pb.ListArtsRequest]) (*connect.Response[pb.ListArtsResponse], error) {
	response, err := a.server.ListArts(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(response), nil
}

// DeleteArt implements the Connect handler interface
func (a *ConnectAdapter) DeleteArt(ctx context.Context, req *connect.Request[pb.DeleteArtRequest]) (*connect.Response[emptypb.Empty], error) {
	_, err := a.server.DeleteArt(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// GetArtUploadUrl implements the Connect handler interface
func (a *ConnectAdapter) GetArtUploadUrl(ctx context.Context, req *connect.Request[pb.GetArtUploadUrlRequest]) (*connect.Response[pb.GetArtUploadUrlResponse], error) {
	response, err := a.server.GetArtUploadUrl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(response), nil
}

// ConfirmArtImageUpload implements the Connect handler interface
func (a *ConnectAdapter) ConfirmArtImageUpload(ctx context.Context, req *connect.Request[pb.ConfirmArtImageUploadRequest]) (*connect.Response[pb.Art], error) {
	art, err := a.server.ConfirmArtImageUpload(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(art), nil
}

// CreateComposition implements the Connect handler interface
func (a *ConnectAdapter) CreateComposition(ctx context.Context, req *connect.Request[pb.CreateCompositionRequest]) (*connect.Response[pb.Composition], error) {
	composition, err := a.server.CreateComposition(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(composition), nil
}

// GetComposition implements the Connect handler interface
func (a *ConnectAdapter) GetComposition(ctx context.Context, req *connect.Request[pb.GetCompositionRequest]) (*connect.Response[pb.Composition], error) {
	composition, err := a.server.GetComposition(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(composition), nil
}

// UpdateComposition implements the Connect handler interface
func (a *ConnectAdapter) UpdateComposition(ctx context.Context, req *connect.Request[pb.UpdateCompositionRequest]) (*connect.Response[pb.Composition], error) {
	composition, err := a.server.UpdateComposition(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(composition), nil
}

// ListCompositions implements the Connect handler interface
func (a *ConnectAdapter) ListCompositions(ctx context.Context, req *connect.Request[pb.ListCompositionsRequest]) (*connect.Response[pb.ListCompositionsResponse], error) {
	response, err := a.server.ListCompositions(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(response), nil
}

// DeleteComposition implements the Connect handler interface
func (a *ConnectAdapter) DeleteComposition(ctx context.Context, req *connect.Request[pb.DeleteCompositionRequest]) (*connect.Response[emptypb.Empty], error) {
	_, err := a.server.DeleteComposition(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}
