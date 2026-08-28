package service

import (
	"context"
	"database/sql"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Damione1/thread-art-generator/core/db/models"
	pbErrors "github.com/Damione1/thread-art-generator/core/errors"
	"github.com/Damione1/thread-art-generator/core/pb"
	"github.com/Damione1/thread-art-generator/core/pbx"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/Damione1/thread-art-generator/core/storage"
	"github.com/bufbuild/protovalidate-go"
	"github.com/pkg/errors"
	"github.com/rs/zerolog/log"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
	"github.com/volatiletech/sqlboiler/v4/queries/qm"
	"go.einride.tech/aip/fieldmask"
	"google.golang.org/genproto/googleapis/rpc/errdetails"
	"google.golang.org/protobuf/types/known/emptypb"
	"google.golang.org/protobuf/types/known/fieldmaskpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func (server *Server) createArt(ctx context.Context, req *pb.CreateArtRequest) (*pb.Art, error) {
	user, err := server.currentUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}
	if err := requireParentIdentity(req.GetParent(), user.ID); err != nil {
		return nil, err
	}
	if user.Role != models.RoleEnumUser {
		return nil, pbErrors.PermissionDeniedError("only users can create art")
	}
	artDb := &models.Art{
		Title:    req.GetArt().GetTitle(),
		AuthorID: user.ID,
		Status:   models.ArtStatusEnumPENDING_IMAGE,
	}

	err = artDb.Insert(ctx, server.config.DB, boil.Infer())
	if err != nil {
		return nil, pbErrors.InternalError("failed to insert art", err)
	}

	return pbx.ArtDbToProto(artDb, server.storage), nil
}

func (server *Server) updateArt(ctx context.Context, req *pb.UpdateArtRequest) (*pb.Art, error) {
	user, err := server.currentUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	artResource, err := resource.ParseResourceName(req.GetArt().GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("art.name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("art.name", errors.New("invalid art resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	// Compare internal user ID with art's user ID from resource name
	if art.UserID != user.ID {
		return nil, pbErrors.PermissionDeniedError("only the author can update the art")
	}

	// Check if the art exists
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(art.ArtID),
		models.ArtWhere.AuthorID.EQ(user.ID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("art not found")
		}
		return nil, pbErrors.InternalError("failed to get art", err)
	}

	cols, err := applyArtUpdateMask(req.GetUpdateMask(), req.GetArt(), artDb)
	if err != nil {
		return nil, err
	}
	if len(cols) > 0 {
		_, err = artDb.Update(ctx, server.config.DB, boil.Whitelist(cols...))
		if err != nil {
			return nil, pbErrors.InternalError("failed to update art", err)
		}
	}

	return pbx.ArtDbToProto(artDb, server.storage), nil
}

func (server *Server) listArts(ctx context.Context, req *pb.ListArtsRequest) (*pb.ListArtsResponse, error) {
	user, err := server.currentUser(ctx)
	if err != nil {
		return nil, err
	}

	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}
	if err := requireParentIdentity(req.GetParent(), user.ID); err != nil {
		return nil, err
	}

	if req.GetPageSize() < 0 {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("page_size", errors.New("page size is negative")),
		})
	}
	pageSize := int(clampPageSize(req.GetPageSize(), 100, 100))

	offset, err := pageOffset(req)
	if err != nil {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("page_token", err),
		})
	}

	orderColumn, dir := parseOrderBy(req.GetOrderBy())

	// Build query mods using internal user ID
	queryMods := []qm.QueryMod{
		models.ArtWhere.AuthorID.EQ(user.ID),
		qm.OrderBy(fmt.Sprintf("%s %s", orderColumn, dir)),
		qm.Limit(pageSize + 1),
		qm.Offset(offset),
	}

	// Query the arts with pagination and sorting
	arts, err := models.Arts(queryMods...).All(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to get arts", err)
	}

	// Check if there are more results
	hasNextPage := false
	if len(arts) > pageSize {
		hasNextPage = true
		arts = arts[:pageSize] // Trim the extra result
	}

	// Convert the arts to protobuf format
	artPbs := make([]*pb.Art, 0, len(arts))
	for _, artDb := range arts {
		artPbs = append(artPbs, pbx.ArtDbToProto(artDb, server.storage))
	}

	return &pb.ListArtsResponse{
		Arts:          artPbs,
		NextPageToken: encodeNextPageToken(req, int32(pageSize), hasNextPage),
	}, nil
}

func parseOrderBy(orderBy string) (column, dir string) {
	column = models.ArtColumns.CreatedAt
	dir = "DESC"
	fields := strings.Fields(strings.TrimSpace(orderBy))
	if len(fields) == 0 {
		return column, dir
	}
	switch fields[0] {
	case "update_time":
		column = models.ArtColumns.UpdatedAt
	default:
		column = models.ArtColumns.CreatedAt
	}
	if len(fields) > 1 && strings.EqualFold(fields[1], "asc") {
		dir = "ASC"
	}
	return column, dir
}

func (server *Server) getArt(ctx context.Context, req *pb.GetArtRequest) (*pb.Art, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	user, err := server.currentUser(ctx)
	if err != nil {
		return nil, err
	}

	artResource, err := resource.ParseResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid art resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	// Compare internal user ID with art's user ID from resource name
	if art.UserID != user.ID {
		return nil, pbErrors.PermissionDeniedError("only the author can get the art")
	}

	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(art.ArtID),
		models.ArtWhere.AuthorID.EQ(user.ID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("art not found")
		}
		return nil, pbErrors.InternalError("failed to get art", err)
	}

	return pbx.ArtDbToProto(artDb, server.storage), nil
}

func (server *Server) deleteArt(ctx context.Context, req *pb.DeleteArtRequest) (*emptypb.Empty, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}

	user, err := server.currentUser(ctx)
	if err != nil {
		return nil, err
	}

	artResource, err := resource.ParseResourceName(req.GetName())
	if err != nil {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	art, ok := artResource.(*resource.Art)
	if !ok {
		violations := []*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid art resource name")),
		}
		return nil, pbErrors.InvalidArgumentError(violations)
	}

	// Compare internal user ID with art's user ID from resource name
	if art.UserID != user.ID {
		return nil, pbErrors.PermissionDeniedError("only the author can delete the art")
	}

	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(art.ArtID),
		models.ArtWhere.AuthorID.EQ(user.ID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, pbErrors.NotFoundError("art not found")
		}
		return nil, pbErrors.InternalError("failed to get art", err)
	}

	_, err = artDb.Delete(ctx, server.config.DB)
	if err != nil {
		return nil, pbErrors.InternalError("failed to delete art", err)
	}

	// Delete the image from the bucket
	if artDb.ImageID.Valid {
		imageKey := resource.ArtImageObjectKey(artDb.AuthorID, artDb.ID, artDb.ImageID.String)
		err = server.storage.GetPublicStorage().Delete(ctx, imageKey)
		if err != nil {
			log.Error().Err(err).Msg(fmt.Sprintf("Failed to delete image %s", artDb.ImageID.String))
			return &emptypb.Empty{}, nil // Don't return a public error if the image deletion fails
		}
	}

	return &emptypb.Empty{}, nil
}

func flattenPresignHeaders(h http.Header) map[string]string {
	out := make(map[string]string, len(h))
	for k, vals := range h {
		if len(vals) > 0 {
			out[k] = vals[0]
		}
	}
	return out
}

const maxArtImageBytes = 10 * 1024 * 1024

func validateUploadedObject(info *storage.ObjectInfo) error {
	if info == nil {
		return pbErrors.FailedPreconditionError("image not found in storage, upload first")
	}
	if info.ContentType != "" && !strings.HasPrefix(info.ContentType, "image/") {
		return pbErrors.FailedPreconditionError("uploaded object is not an image")
	}
	if info.Size > maxArtImageBytes {
		return pbErrors.FailedPreconditionError("uploaded object exceeds 10MB")
	}
	return nil
}

func (server *Server) loadOwnedArt(ctx context.Context, name, action string) (*models.Art, *models.User, error) {
	user, err := server.currentUser(ctx)
	if err != nil {
		return nil, nil, err
	}
	parsed, err := resource.ParseResourceName(name)
	if err != nil {
		return nil, nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid resource name")),
		})
	}
	artRes, ok := parsed.(*resource.Art)
	if !ok {
		return nil, nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("name", errors.New("invalid art resource name")),
		})
	}
	if artRes.UserID != user.ID {
		return nil, nil, pbErrors.PermissionDeniedError("only the author can " + action)
	}
	artDb, err := models.Arts(
		models.ArtWhere.ID.EQ(artRes.ArtID),
		models.ArtWhere.AuthorID.EQ(user.ID),
	).One(ctx, server.config.DB)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, pbErrors.NotFoundError("art not found")
		}
		return nil, nil, pbErrors.InternalError("failed to get art", err)
	}
	return artDb, user, nil
}

func (server *Server) startArtUpload(ctx context.Context, req *pb.StartArtUploadRequest) (*pb.StartArtUploadResponse, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}
	artDb, user, err := server.loadOwnedArt(ctx, req.GetName(), "upload")
	if err != nil {
		return nil, err
	}
	if artDb.Status != models.ArtStatusEnumPENDING_IMAGE {
		return nil, pbErrors.FailedPreconditionError("art is not awaiting an image")
	}
	key := resource.ArtOriginalObjectKey(user.ID, artDb.ID)
	resp, err := presignArtOriginal(ctx, server.storage.GetPublicStorage().Bucket(), key, req.GetContentType())
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (server *Server) completeArtUpload(ctx context.Context, req *pb.CompleteArtUploadRequest) (*pb.Art, error) {
	if err := protovalidate.Validate(req); err != nil {
		return nil, pbErrors.ConvertProtoValidateError(err)
	}
	artDb, user, err := server.loadOwnedArt(ctx, req.GetName(), "complete upload")
	if err != nil {
		return nil, err
	}
	key := resource.ArtOriginalObjectKey(user.ID, artDb.ID)
	if err := headUploadedOriginal(ctx, server.storage.GetPublicStorage().Bucket(), key); err != nil {
		return nil, err
	}
	artDb.Status = models.ArtStatusEnumCOMPLETE
	artDb.ImageID = null.StringFrom(artDb.ID)
	_, err = artDb.Update(ctx, server.config.DB, boil.Whitelist(models.ArtColumns.Status, models.ArtColumns.ImageID))
	if err != nil {
		return nil, pbErrors.InternalError("failed to update art status", err)
	}
	return pbx.ArtDbToProto(artDb, server.storage), nil
}

func presignArtOriginal(ctx context.Context, bucket storage.Bucket, key, contentType string) (*pb.StartArtUploadResponse, error) {
	presign, err := bucket.PresignPut(ctx, key, storage.PresignPutOptions{
		ContentType: contentType,
		TTL:         10 * time.Minute,
	})
	if err != nil {
		return nil, pbErrors.InternalError("failed to presign upload", err)
	}
	return &pb.StartArtUploadResponse{
		UploadUrl: presign.URL,
		Method:    presign.Method,
		Headers:   flattenPresignHeaders(presign.Headers),
		ExpiresAt: timestamppb.New(presign.Expires),
	}, nil
}

func headUploadedOriginal(ctx context.Context, bucket storage.Bucket, key string) error {
	info, err := bucket.Head(ctx, key)
	if err != nil {
		return pbErrors.FailedPreconditionError("image not found in storage, upload first")
	}
	return validateUploadedObject(info)
}

func requireParentIdentity(parent, userID string) error {
	res, err := resource.ParseResourceName(parent)
	if err != nil {
		return pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("parent", errors.New("invalid resource name")),
		})
	}
	userRes, ok := res.(*resource.User)
	if !ok {
		return pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("parent", errors.New("invalid user resource name")),
		})
	}
	if userRes.ID != userID {
		return pbErrors.PermissionDeniedError("parent does not match authenticated user")
	}
	return nil
}

func applyArtUpdateMask(mask *fieldmaskpb.FieldMask, src *pb.Art, dst *models.Art) ([]string, error) {
	if src == nil || dst == nil {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("art", errors.New("art is required")),
		})
	}
	paths := mask.GetPaths()
	if fieldmask.IsFullReplacement(mask) || len(paths) == 0 {
		paths = []string{"title"}
	} else if err := fieldmask.Validate(mask, src); err != nil {
		return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
			pbErrors.FieldViolation("update_mask", err),
		})
	}

	cols := make([]string, 0, 2)
	seenTitle := false
	for _, path := range paths {
		switch path {
		case "title":
			if seenTitle {
				continue
			}
			seenTitle = true
			dst.Title = src.GetTitle()
			cols = append(cols, models.ArtColumns.Title)
		case "name", "image_url", "status", "author", "create_time", "update_time":
			return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
				pbErrors.FieldViolation("update_mask", fmt.Errorf("field %q is not updatable", path)),
			})
		default:
			return nil, pbErrors.InvalidArgumentError([]*errdetails.BadRequest_FieldViolation{
				pbErrors.FieldViolation("update_mask", fmt.Errorf("unknown field %q", path)),
			})
		}
	}
	return cols, nil
}
