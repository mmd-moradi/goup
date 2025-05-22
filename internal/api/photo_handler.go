package api

import (
	"encoding/json"
	"io"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/mmd-moradi/goup/internal/middleware"
	"github.com/mmd-moradi/goup/internal/service"
	"github.com/mmd-moradi/goup/pkg/apperrors"
	"github.com/mmd-moradi/goup/pkg/response"
)

type PhotoHandler struct {
	photoService *service.PhotoService
}

func NewPhotoHandler(photoService *service.PhotoService) *PhotoHandler {
	return &PhotoHandler{
		photoService: photoService,
	}
}

// Upload handles photo upload
// @Summary Upload a new photo
// @Description Upload a new photo with metadata
// @Tags photos
// @Accept multipart/form-data
// @Produce json
// @Param title formData string true "Photo title"
// @Param description formData string false "Photo description"
// @Param file formData file true "Photo file to upload"
// @Security Bearer
// @Success 201 {object} response.Response{data=service.PhotoResponse} "Photo uploaded successfully"
// @Failure 400 {object} response.Response{error=response.ErrorInfo} "Invalid request payload"
// @Failure 401 {object} response.Response{error=response.ErrorInfo} "User not authenticated"
// @Failure 500 {object} response.Response{error=response.ErrorInfo} "Internal server error"
// @Router /photos [post]
func (h *PhotoHandler) Upload(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, 10*1024*1024)
	if err := r.ParseMultipartForm(10 << 20); err != nil {
		response.Error(w, apperrors.NewWithFormat(apperrors.BadRequest, "failed to parse form: %v", err))
	}

	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	title := r.FormValue("title")
	description := r.FormValue("description")
	file, header, err := r.FormFile("file")
	if err != nil {
		response.Error(w, apperrors.NewWithFormat(apperrors.BadRequest, "failed to get file: %v", err))
		return
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		response.Error(w, apperrors.NewWithFormat(apperrors.InternalServer, "failed to read file: %v", err))
		return
	}

	input := service.PhotoUploadInput{
		Title:       title,
		Description: description,
		FileName:    header.Filename,
		FileSize:    header.Size,
		ContentType: header.Header.Get("Content-Type"),
	}

	photo, err := h.photoService.UploadPhoto(r.Context(), input, userID, data)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusCreated, photo)
}

// GetByID handles getting a photo by ID
// @Summary Get a photo by ID
// @Description Get a photo by its ID
// @Tags photos
// @Produce json
// @Param id path string true "Photo ID"
// @Security Bearer
// @Success 200 {object} response.Response{data=service.PhotoResponse} "Photo retrieved successfully"
// @Failure 400 {object} response.Response{error=response.ErrorInfo} "Invalid photo ID"
// @Failure 401 {object} response.Response{error=response.ErrorInfo} "User not authenticated"
// @Failure 403 {object} response.Response{error=response.ErrorInfo} "User doesn't have access to the photo"
// @Failure 404 {object} response.Response{error=response.ErrorInfo} "Photo not found"
// @Failure 500 {object} response.Response{error=response.ErrorInfo} "Internal server error"
// @Router /photos/{id} [get]
func (h *PhotoHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	photoIDStr := chi.URLParam(r, "id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		response.Error(w, apperrors.New(apperrors.BadRequest, "invalid photo ID"))
	}

	photo, err := h.photoService.GetPhotoByID(r.Context(), photoID, userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, photo)
}

// List handles listing photos for the current user with pagination
// @Summary List user photos
// @Description Get a paginated list of photos for the authenticated user
// @Tags photos
// @Produce json
// @Param page query int false "Page number (default: 1)"
// @Param page_size query int false "Page size (default: 10, max: 100)"
// @Security Bearer
// @Success 200 {object} response.Response{data=service.PhotosResponse} "Photos retrieved successfully"
// @Failure 401 {object} response.Response{error=response.ErrorInfo} "User not authenticated"
// @Failure 500 {object} response.Response{error=response.ErrorInfo} "Internal server error"
// @Router /photos [get]
func (h *PhotoHandler) List(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}

	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("page_size")

	page := 1
	pageSize := 10

	if pageStr != "" {
		pageInt, err := strconv.Atoi(pageStr)
		if err == nil && page > 0 {
			page = pageInt
		}
	}

	if pageSizeStr != "" {
		pageSizeInt, err := strconv.Atoi(pageSizeStr)
		if err == nil && pageSize > 0 && pageSizeInt <= 100 {
			pageSize = pageSizeInt
		}
	}

	listPhotos, err := h.photoService.GetPhotosByID(r.Context(), userID, page, pageSize)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, listPhotos)

}

// Update handles updating a photo's metadata
// @Summary Update photo metadata
// @Description Update the title and description of a photo
// @Tags photos
// @Accept json
// @Produce json
// @Param id path string true "Photo ID"
// @Param input body service.PhotoUpdateInput true "Photo update information"
// @Security Bearer
// @Success 200 {object} response.Response{data=service.PhotoResponse} "Photo updated successfully"
// @Failure 400 {object} response.Response{error=response.ErrorInfo} "Invalid request payload"
// @Failure 401 {object} response.Response{error=response.ErrorInfo} "User not authenticated"
// @Failure 403 {object} response.Response{error=response.ErrorInfo} "User doesn't have access to the photo"
// @Failure 404 {object} response.Response{error=response.ErrorInfo} "Photo not found"
// @Failure 500 {object} response.Response{error=response.ErrorInfo} "Internal server error"
// @Router /photos/{id} [put]
func (h *PhotoHandler) Update(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	photoIDStr := chi.URLParam(r, "id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		response.Error(w, apperrors.New(apperrors.BadRequest, "invalid photo ID"))
		return
	}

	var input service.PhotoUpdateInput
	err = json.NewDecoder(r.Body).Decode(&input)
	if err != nil {
		response.Error(w, apperrors.New(apperrors.BadRequest, "invalid request payload"))
		return
	}

	updatedPhoto, err := h.photoService.UpdatePhoto(r.Context(), photoID, input, userID)
	if err != nil {
		response.Error(w, err)
		return
	}

	response.JSON(w, http.StatusOK, updatedPhoto)
}

// Delete handles deleting a photo
// @Summary Delete a photo
// @Description Delete a photo by its ID
// @Tags photos
// @Param id path string true "Photo ID"
// @Security Bearer
// @Success 204 "Photo deleted successfully"
// @Failure 400 {object} response.Response{error=response.ErrorInfo} "Invalid photo ID"
// @Failure 401 {object} response.Response{error=response.ErrorInfo} "User not authenticated"
// @Failure 403 {object} response.Response{error=response.ErrorInfo} "User doesn't have access to the photo"
// @Failure 404 {object} response.Response{error=response.ErrorInfo} "Photo not found"
// @Failure 500 {object} response.Response{error=response.ErrorInfo} "Internal server error"
// @Router /photos/{id} [delete]
func (h *PhotoHandler) Delete(w http.ResponseWriter, r *http.Request) {
	userID, err := middleware.GetUserID(r.Context())
	if err != nil {
		response.Error(w, err)
		return
	}
	photoIDStr := chi.URLParam(r, "id")
	photoID, err := uuid.Parse(photoIDStr)
	if err != nil {
		response.Error(w, apperrors.New(apperrors.BadRequest, "invalid photo ID"))
	}

	err = h.photoService.DeletePhoto(r.Context(), photoID, userID)
	if err != nil {
		response.Error(w, err)
		return
	}
	response.NoContent(w)
}

func (h *PhotoHandler) RegisterRoutes(r chi.Router, authMiddleware func(next http.Handler) http.Handler) {
	r.Group(func(r chi.Router) {
		r.Use(authMiddleware)
		r.Post("/", h.Upload)
		r.Get("/", h.List)
		r.Get("/{id}", h.GetByID)
		r.Put("/{id}", h.Update)
		r.Delete("/{id}", h.Delete)
	})
}
