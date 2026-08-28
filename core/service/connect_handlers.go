package service

import (
	"context"
	"fmt"

	"connectrpc.com/connect"
	"github.com/Damione1/thread-art-generator/core/pb"
	"google.golang.org/protobuf/types/known/emptypb"
)

func (s *Server) UpdateUser(ctx context.Context, req *connect.Request[pb.UpdateUserRequest]) (*connect.Response[pb.User], error) {
	user, err := s.updateUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(user), nil
}

func (s *Server) GetUser(ctx context.Context, req *connect.Request[pb.GetUserRequest]) (*connect.Response[pb.User], error) {
	user, err := s.getUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(user), nil
}

func (s *Server) ListUsers(ctx context.Context, req *connect.Request[pb.ListUsersRequest]) (*connect.Response[pb.ListUsersResponse], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("pb.ArtGeneratorService.ListUsers is not implemented"))
}

func (s *Server) DeleteUser(ctx context.Context, req *connect.Request[pb.DeleteUserRequest]) (*connect.Response[emptypb.Empty], error) {
	return nil, connect.NewError(connect.CodeUnimplemented, fmt.Errorf("pb.ArtGeneratorService.DeleteUser is not implemented"))
}

func (s *Server) GetCurrentUser(ctx context.Context, req *connect.Request[pb.GetCurrentUserRequest]) (*connect.Response[pb.User], error) {
	user, err := s.getCurrentUser(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(user), nil
}

func (s *Server) SyncUserFromFirebase(ctx context.Context, req *connect.Request[pb.SyncUserFromFirebaseRequest]) (*connect.Response[pb.User], error) {
	if !s.validateInternalAPIKeyFromHeaders(req.Header()) {
		return nil, connect.NewError(connect.CodeUnauthenticated, fmt.Errorf("invalid internal API key"))
	}
	user, err := s.syncUserFromFirebase(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(user), nil
}

func (s *Server) CreateArt(ctx context.Context, req *connect.Request[pb.CreateArtRequest]) (*connect.Response[pb.Art], error) {
	art, err := s.createArt(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(art), nil
}

func (s *Server) GetArt(ctx context.Context, req *connect.Request[pb.GetArtRequest]) (*connect.Response[pb.Art], error) {
	art, err := s.getArt(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(art), nil
}

func (s *Server) UpdateArt(ctx context.Context, req *connect.Request[pb.UpdateArtRequest]) (*connect.Response[pb.Art], error) {
	art, err := s.updateArt(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(art), nil
}

func (s *Server) ListArts(ctx context.Context, req *connect.Request[pb.ListArtsRequest]) (*connect.Response[pb.ListArtsResponse], error) {
	resp, err := s.listArts(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Server) DeleteArt(ctx context.Context, req *connect.Request[pb.DeleteArtRequest]) (*connect.Response[emptypb.Empty], error) {
	_, err := s.deleteArt(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func (s *Server) GetArtUploadUrl(ctx context.Context, req *connect.Request[pb.GetArtUploadUrlRequest]) (*connect.Response[pb.GetArtUploadUrlResponse], error) {
	resp, err := s.getArtUploadUrl(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Server) ConfirmArtImageUpload(ctx context.Context, req *connect.Request[pb.ConfirmArtImageUploadRequest]) (*connect.Response[pb.Art], error) {
	art, err := s.confirmArtImageUpload(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(art), nil
}

func (s *Server) StartArtUpload(ctx context.Context, req *connect.Request[pb.StartArtUploadRequest]) (*connect.Response[pb.StartArtUploadResponse], error) {
	resp, err := s.startArtUpload(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Server) CompleteArtUpload(ctx context.Context, req *connect.Request[pb.CompleteArtUploadRequest]) (*connect.Response[pb.Art], error) {
	art, err := s.completeArtUpload(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(art), nil
}

func (s *Server) CreateComposition(ctx context.Context, req *connect.Request[pb.CreateCompositionRequest]) (*connect.Response[pb.Composition], error) {
	composition, err := s.createComposition(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(composition), nil
}

func (s *Server) GetComposition(ctx context.Context, req *connect.Request[pb.GetCompositionRequest]) (*connect.Response[pb.Composition], error) {
	composition, err := s.getComposition(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(composition), nil
}

func (s *Server) UpdateComposition(ctx context.Context, req *connect.Request[pb.UpdateCompositionRequest]) (*connect.Response[pb.Composition], error) {
	composition, err := s.updateComposition(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(composition), nil
}

func (s *Server) ListCompositions(ctx context.Context, req *connect.Request[pb.ListCompositionsRequest]) (*connect.Response[pb.ListCompositionsResponse], error) {
	resp, err := s.listCompositions(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(resp), nil
}

func (s *Server) DeleteComposition(ctx context.Context, req *connect.Request[pb.DeleteCompositionRequest]) (*connect.Response[emptypb.Empty], error) {
	_, err := s.deleteComposition(ctx, req.Msg)
	if err != nil {
		return nil, err
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}
