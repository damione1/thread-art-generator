package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/rs/zerolog/log"

	"github.com/Damione1/thread-art-generator/client/internal/middleware"
	"github.com/Damione1/thread-art-generator/client/internal/services"
)

// UploadHandler handles file upload operations
type UploadHandler struct {
	generatorService *services.GeneratorService
}

// NewUploadHandler creates a new upload handler
func NewUploadHandler(generatorService *services.GeneratorService) *UploadHandler {
	return &UploadHandler{
		generatorService: generatorService,
	}
}

// GenerateUploadURLRequest represents the JSON request for upload URL generation
type GenerateUploadURLRequest struct {
	ArtID       string `json:"artId"`
	ContentType string `json:"contentType"`
	FileSize    int64  `json:"fileSize"`
	Filename    string `json:"filename,omitempty"`
}

// GenerateUploadURLResponse represents the JSON response for upload URL generation
type GenerateUploadURLResponse struct {
	UploadURL     string `json:"uploadUrl"`
	StoragePath   string `json:"storagePath"`
	ExpiresAt     string `json:"expiresAt"`
	MaxFileSize   int64  `json:"maxFileSize"`
}

// ErrorResponse represents an error response
type ErrorResponse struct {
	Error   string `json:"error"`
	Message string `json:"message,omitempty"`
}

// GenerateUploadURL generates a signed URL for file uploads
func (h *UploadHandler) GenerateUploadURL(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	user, ok := middleware.UserFromContext(r.Context())
	if !ok || user == nil {
		log.Error().Msg("User not found in context for upload URL generation")
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", "")
		return
	}

	// Parse JSON request
	var req GenerateUploadURLRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Error().Err(err).Msg("Failed to parse upload URL request")
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid request body", err.Error())
		return
	}

	// Validate required fields
	if req.ArtID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing required field: artId", "")
		return
	}
	if req.ContentType == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing required field: contentType", "")
		return
	}
	if req.FileSize <= 0 {
		h.writeErrorResponse(w, http.StatusBadRequest, "Invalid file size", "File size must be greater than 0")
		return
	}

	log.Info().
		Str("user_id", user.ID).
		Str("art_id", req.ArtID).
		Str("content_type", req.ContentType).
		Int64("file_size", req.FileSize).
		Msg("Processing upload URL request")

	// Call generator service (same pattern as other endpoints)
	serviceReq := services.StorageUploadURLRequest{
		ArtID:       req.ArtID,
		ContentType: req.ContentType,
		FileSize:    req.FileSize,
		Filename:    req.Filename,
	}

	response, err := h.generatorService.GenerateUploadURL(r.Context(), r, user.ID, serviceReq)
	if err != nil {
		log.Error().Err(err).Msg("Storage service failed to generate upload URL")
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate upload URL", err.Error())
		return
	}

	// Return successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	jsonResponse := GenerateUploadURLResponse{
		UploadURL:     response.UploadURL,
		StoragePath:   response.StoragePath,
		ExpiresAt:     response.ExpiresAt,
		MaxFileSize:   response.MaxFileSize,
	}

	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		log.Error().Err(err).Msg("Failed to encode upload URL response")
		return
	}

	log.Info().
		Str("user_id", user.ID).
		Str("art_id", req.ArtID).
		Str("storage_path", response.StoragePath).
		Msg("Upload URL generated and returned successfully")
}

// GenerateDownloadURLRequest represents the JSON request for download URL generation
type GenerateDownloadURLRequest struct {
	ArtID    string  `json:"artId"`
	FilePath *string `json:"filePath,omitempty"`
}

// GenerateDownloadURLResponse represents the JSON response for download URL generation
type GenerateDownloadURLResponse struct {
	DownloadURL string `json:"downloadUrl"`
	StoragePath string `json:"storagePath"`
	ExpiresAt   string `json:"expiresAt"`
}

// GenerateDownloadURL generates a signed URL for file downloads
func (h *UploadHandler) GenerateDownloadURL(w http.ResponseWriter, r *http.Request) {
	// Get user from context (set by auth middleware)
	user, ok := middleware.UserFromContext(r.Context())
	if !ok || user == nil {
		log.Error().Msg("User not found in context for download URL generation")
		h.writeErrorResponse(w, http.StatusUnauthorized, "Authentication required", "")
		return
	}

	// Get art ID from URL path
	artID := chi.URLParam(r, "artId")
	if artID == "" {
		h.writeErrorResponse(w, http.StatusBadRequest, "Missing art ID in URL path", "")
		return
	}

	// Parse optional JSON request for file path
	var req GenerateDownloadURLRequest
	req.ArtID = artID
	
	if r.Body != nil {
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			// If body parsing fails, it's okay - we'll use defaults
			log.Debug().Err(err).Msg("No JSON body provided for download URL request, using defaults")
		}
	}

	log.Info().
		Str("user_id", user.ID).
		Str("art_id", req.ArtID).
		Msg("Processing download URL request")

	// Call generator service (same pattern as other endpoints)
	serviceReq := services.StorageDownloadURLRequest{
		ArtID:    req.ArtID,
		FilePath: req.FilePath,
	}

	response, err := h.generatorService.GenerateDownloadURL(r.Context(), r, user.ID, serviceReq)
	if err != nil {
		log.Error().Err(err).Msg("Storage service failed to generate download URL")
		h.writeErrorResponse(w, http.StatusInternalServerError, "Failed to generate download URL", err.Error())
		return
	}

	// Return successful response
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	
	jsonResponse := GenerateDownloadURLResponse{
		DownloadURL: response.DownloadURL,
		StoragePath: response.StoragePath,
		ExpiresAt:   response.ExpiresAt,
	}

	if err := json.NewEncoder(w).Encode(jsonResponse); err != nil {
		log.Error().Err(err).Msg("Failed to encode download URL response")
		return
	}

	log.Info().
		Str("user_id", user.ID).
		Str("art_id", req.ArtID).
		Str("storage_path", response.StoragePath).
		Msg("Download URL generated and returned successfully")
}

// writeErrorResponse writes a standardized error response
func (h *UploadHandler) writeErrorResponse(w http.ResponseWriter, statusCode int, error, message string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	
	response := ErrorResponse{
		Error:   error,
		Message: message,
	}
	
	if err := json.NewEncoder(w).Encode(response); err != nil {
		log.Error().Err(err).Msg("Failed to encode error response")
	}
}