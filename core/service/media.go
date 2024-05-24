package service

import (
	"context"
	"database/sql"
	"fmt"
	"mime"
	"net/http"
	"strings"

	"github.com/Damione1/thread-art-generator/core/middleware"
	"github.com/Damione1/thread-art-generator/core/resource"
	"github.com/Damione1/thread-art-generator/core/token"
	"github.com/friendsofgo/errors"
	validation "github.com/go-ozzo/ozzo-validation"
)

/*
* The code below is the implementation of the HandleBinaryFileUpload function.
* This function is responsible for handling image upload. It validates the upload request, checks if the user is authorized to upload the file in a valid ressource, and updates the database with the uploaded file.
* The code also includes the implementation of the Resource interface and the Art struct.
* The ressource interface defines two methods: Validate and UpdateDB. They are meant to be implemented by the specific resource types if needed in the future.
 */

// Constants
const (
	MaxUploadSize     = 10 * 1024 * 1024 // 10MB
	MultipartFormData = "multipart/form-data"
)

// Validate upload request
func validateUploadRequest(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")

	// Check if the content type is multipart/form-data
	if !strings.HasPrefix(contentType, MultipartFormData) {
		return errors.New(fmt.Sprintf("Content type is not %s", MultipartFormData))
	}

	err := r.ParseMultipartForm(MaxUploadSize)
	if err != nil {
		return errors.New("The uploaded file is too big. Please choose an appropriate size")
	}

	resourceName := r.FormValue("name")
	if resourceName == "" {
		return errors.New("Resource name is required")
	}

	if r.MultipartForm.File["file"] == nil {
		return errors.New("File is required")
	}

	return nil
}

func validateGetMediaUploadUrlRequest(r *http.Request) error {
	return validation.Errors{
		"name": validation.Validate(r.FormValue("name"), validation.Required, validation.Length(5, 20)),
	}.Filter()
}

// Validate content type and extension of the uploaded file
func validateFile(contentType, extension string) error {
	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return errors.New("Invalid content type")
	}
	if mediaType != "image/jpeg" && mediaType != "image/png" {
		return errors.New("Invalid content type")
	}
	if extension != ".jpeg" && extension != ".png" {
		return errors.New("Invalid extension")
	}
	return nil
}

// HandleBinaryFileUpload handles binary file upload
func HandleBinaryFileUpload(server *Server) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		// Get the authenticated user payload from the context
		userPayload := middleware.FromAdminContext(ctx).UserPayload

		// Validate the upload request
		if err := validateUploadRequest(r); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		resourceName := r.FormValue("name")

		uploadedResource, err := newUploadedResource(ctx, server.config.DB, resourceName)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if err := uploadedResource.Validate(userPayload); err != nil {
			http.Error(w, err.Error(), http.StatusForbidden)
			return
		}

		file, fileHeader, err := r.FormFile("file")
		if err != nil {
			http.Error(w, "Failed to get uploaded file", http.StatusBadRequest)
			return
		}
		defer file.Close()

		if fileHeader.Size > MaxUploadSize {
			http.Error(w, "The uploaded file is too large. Please choose a file under 10MB", http.StatusBadRequest)
			return
		}

		contentType := fileHeader.Header.Get("Content-Type")
		extension := fileHeader.Filename[strings.LastIndex(fileHeader.Filename, "."):]
		if err := validateFile(contentType, extension); err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		fileKey := fmt.Sprintf("%s%s", resourceName, extension)

		writer, err := server.bucket.NewWriter(ctx, fileKey, nil)
		if err != nil {
			http.Error(w, "Failed to create a new writer", http.StatusInternalServerError)
			return
		}
		defer writer.Close()

		_, err = writer.ReadFrom(file)
		if err != nil {
			http.Error(w, "Failed to write to the bucket", http.StatusInternalServerError)
			return
		}

		if err := uploadedResource.UpdateDB(fileKey); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}

// Add a case for 'anotherResource' in the factory function
func newUploadedResource(ctx context.Context, db *sql.DB, resourceName string) (Resource, error) {
	rp, err := resource.NewResourceParser(resourceName)
	if err != nil {
		return nil, err
	}

	res, err := rp.Parse()
	if err != nil {
		return nil, err
	}

	switch ressourceType := res.(type) {
	case *resource.Art:
		return &Art{
			Ctx:    ctx,
			Db:     db,
			ArtID:  ressourceType.ArtID,
			UserID: ressourceType.UserID,
		}, nil
	default:
		return nil, errors.New("Unknown resource type")
	}
}

type Resource interface {
	Validate(userPayload *token.Payload) error
	UpdateDB(fileKey string) error
}
