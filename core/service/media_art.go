package service

import (
	"context"
	"database/sql"

	"github.com/Damione1/thread-art-generator/core/db/models"
	"github.com/Damione1/thread-art-generator/core/token"
	"github.com/friendsofgo/errors"
	"github.com/volatiletech/null/v8"
	"github.com/volatiletech/sqlboiler/v4/boil"
)

type Art struct {
	Ctx    context.Context
	Db     *sql.DB
	ArtID  string
	UserID string
	artDb  *models.Art
}

// Authorize checks whether the current user is authorized to modify this resource
func (a *Art) Validate(userPayload *token.Payload) error {
	var err error
	if a.UserID != userPayload.UserID {
		return errors.New("Only the author can upload the image")
	}

	a.artDb, err = models.Arts(
		models.ArtWhere.ID.EQ(a.ArtID),
		models.ArtWhere.AuthorID.EQ(a.UserID),
	).One(a.Ctx, a.Db)
	if err != nil {
		return errors.New("Failed to get art")
	}

	// if a.artDb.ImageID.Valid {
	// 	return errors.New("The art already has an image")
	// }
	return nil
}

// UpdateDB updates the resource in the database
func (a *Art) UpdateDB(fileKey string) error {
	a.artDb.ImageID = null.NewString(fileKey, true)
	if _, err := a.artDb.Update(a.Ctx, a.Db, boil.Infer()); err != nil {
		return errors.New("Failed to update the art in the database")
	}

	return nil
}
