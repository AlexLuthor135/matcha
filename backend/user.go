package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"goji.io/pat"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const maxUploadSize = 5 << 20
const maxUserPhotos = 5
const maxPhotoUploadRequestSize = maxUploadSize*maxUserPhotos + (1 << 20)
const defaultProfileFeedLimit = 20
const maxProfileFeedLimit = 50

var errPhotoLimitReached = errors.New("photo limit reached")

var allowedImageContentTypes = map[string]struct{}{
	"image/jpeg": {},
	"image/png":  {},
	"image/gif":  {},
	"image/webp": {},
}

var allowedImageExtensions = map[string]struct{}{
	".jpg":  {},
	".jpeg": {},
	".png":  {},
	".gif":  {},
	".webp": {},
}

type storedUpload struct {
	FilePath  string
	PublicURL string
}

type UpdateBioRequest struct {
	Gender      string   `json:"gender"`
	Preferences string   `json:"preferences"`
	Bio         string   `json:"bio"`
	Interests   []string `json:"interests"`
}

type BioResponse struct {
	Gender      string          `json:"gender"`
	Preferences string          `json:"preferences"`
	Bio         string          `json:"bio"`
	Interests   []string        `json:"interests"`
	UserName    string          `json:"userName"`
	FirstName   string          `json:"firstName"`
	LastName    string          `json:"lastName"`
	Email       string          `json:"email"`
	Avatar      string          `json:"avatar"`
	Photos      []PhotoResponse `json:"photos"`
}

type PublicProfileResponse struct {
	ID          uint            `json:"id"`
	UserName    string          `json:"userName"`
	FirstName   string          `json:"firstName"`
	Gender      string          `json:"gender"`
	Preferences string          `json:"preferences"`
	Bio         string          `json:"bio"`
	Interests   []string        `json:"interests"`
	Avatar      string          `json:"avatar"`
	Photos      []PhotoResponse `json:"photos"`
}

type ProfileFeedResponse struct {
	Profiles []PublicProfileResponse `json:"profiles"`
	Count    int                     `json:"count"`
}

type PhotoResponse struct {
	ID  uint   `json:"id"`
	URL string `json:"url"`
}

type UpdateUserRequest struct {
	UserName  string `json:"userName"`
	FirstName string `json:"firstName"`
	LastName  string `json:"lastName"`
	Email     string `json:"email"`
}

type UpdatePasswordRequest struct {
	OldPassword string `json:"old_password"`
	NewPassword string `json:"new_password"`
}

func createBio(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}
	var req UpdateBioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	isCompleted := strings.TrimSpace(req.Gender) != "" &&
		strings.TrimSpace(req.Preferences) != "" &&
		strings.TrimSpace(req.Bio) != "" &&
		len(req.Interests) > 0

	result := DB.Model(&User{}).Where("id = ?", userID).Updates(User{
		Gender:      req.Gender,
		Bio:         req.Bio,
		Interests:   req.Interests,
		Preferences: req.Preferences,
		IsCompleted: isCompleted,
	})

	if result.Error != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profile updated successfully",
	})
}

func updateBio(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}
	var req UpdateBioRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	result := DB.Model(&User{}).Where("id = ?", userID).Updates(User{
		Gender:      req.Gender,
		Bio:         req.Bio,
		Interests:   req.Interests,
		Preferences: req.Preferences,
	})

	if result.Error != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profile updated successfully",
	})
}

func getBio(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}

	var user User
	result := DB.Preload("Photos").Omit("password").First(&user, userID)
	if result.Error != nil {
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve profile", http.StatusInternalServerError)
		return
	}

	response := BioResponse{
		Gender:      user.Gender,
		Preferences: user.Preferences,
		Bio:         user.Bio,
		Interests:   user.Interests,
		UserName:    user.UserName,
		FirstName:   user.FirstName,
		LastName:    user.LastName,
		Email:       user.Email,
		Avatar:      user.Avatar,
		Photos:      buildPhotoResponses(user.Photos),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func getProfileFeed(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}

	limit, err := parseProfileFeedLimit(r.URL.Query().Get("limit"))
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	var currentUser User
	if err := DB.Select("id", "gender", "preferences").First(&currentUser, userID).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve current user", http.StatusInternalServerError)
		return
	}

	query := DB.Preload("Photos").
		Where("id <> ? AND is_completed = ?", userID, true)

	if preference := strings.TrimSpace(currentUser.Preferences); preference != "" {
		query = query.Where("gender = ?", preference)
	}
	if gender := strings.TrimSpace(currentUser.Gender); gender != "" {
		query = query.Where("preferences = ?", gender)
	}

	var users []User
	if err := query.Order("RANDOM()").Limit(limit).Find(&users).Error; err != nil {
		http.Error(w, "Failed to retrieve profiles", http.StatusInternalServerError)
		return
	}

	profiles := make([]PublicProfileResponse, 0, len(users))
	for _, user := range users {
		profiles = append(profiles, buildPublicProfileResponse(user))
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(ProfileFeedResponse{
		Profiles: profiles,
		Count:    len(profiles),
	}); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func updateUser(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}
	var req UpdateUserRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	result := DB.Model(&User{}).Where("id = ?", userID).Updates(User{
		UserName:  req.UserName,
		FirstName: req.FirstName,
		LastName:  req.LastName,
		Email:     req.Email,
	})

	if result.Error != nil {
		http.Error(w, "Failed to update profile", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Profile updated successfully",
	})
}

func updatePassword(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}
	var req UpdatePasswordRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}
	if req.OldPassword == "" || req.NewPassword == "" {
		http.Error(w, "Both old and new passwords are required", http.StatusBadRequest)
		return
	}
	var user User
	if err := DB.Select("password").First(&user, userID).Error; err != nil {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}
	err := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(req.OldPassword))
	if err != nil {
		http.Error(w, "Incorrect current password", http.StatusUnauthorized)
		return
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		http.Error(w, "Error processing new password", http.StatusInternalServerError)
		return
	}
	result := DB.Model(&User{}).Where("id = ?", userID).Update("password", string(hashedPassword))
	if result.Error != nil {
		http.Error(w, "Failed to update password", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message": "Password updated successfully",
	})
}

func uploadAvatar(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}

	upload, ok := storeUploadedImage(w, r, userID, "avatar", "avatars")
	if !ok {
		return
	}

	oldAvatar, err := replaceUserAvatar(userID, upload.PublicURL)
	if err != nil {
		_ = os.Remove(upload.FilePath)
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "User not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to upload avatar", http.StatusInternalServerError)
		return
	}

	if oldAvatar != "" && oldAvatar != upload.PublicURL {
		if err := removeStoredUpload(oldAvatar); err != nil {
			log.Printf("failed to remove old avatar for user %d: %v", userID, err)
		}
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Avatar uploaded successfully",
		"avatar_url": upload.PublicURL,
	})
}

func replaceUserAvatar(userID uint, publicURL string) (string, error) {
	var oldAvatar string

	err := DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id", "avatar").First(&user, userID).Error; err != nil {
			return err
		}

		oldAvatar = user.Avatar
		return tx.Model(&User{}).Where("id = ?", userID).Update("avatar", publicURL).Error
	})

	return oldAvatar, err
}

func uploadPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}

	if err := ensurePhotoLimitAvailable(DB, userID); err != nil {
		writePhotoUploadError(w, err)
		return
	}

	var existingPhotoCount int64
	if err := DB.Model(&Photo{}).Where("user_id = ?", userID).Count(&existingPhotoCount).Error; err != nil {
		http.Error(w, "Failed to check uploaded photos", http.StatusInternalServerError)
		return
	}

	remainingSlots := maxUserPhotos - int(existingPhotoCount)
	if remainingSlots <= 0 {
		http.Error(w, "You can upload up to 5 photos", http.StatusBadRequest)
		return
	}

	uploads, ok := storeUploadedImages(w, r, userID, "photos", "photos", remainingSlots)
	if !ok {
		return
	}

	photos := make([]Photo, 0, len(uploads))
	for _, upload := range uploads {
		photos = append(photos, Photo{
			UserID: userID,
			URL:    upload.PublicURL,
		})
	}

	if err := DB.Create(&photos).Error; err != nil {
		for _, upload := range uploads {
			_ = os.Remove(upload.FilePath)
		}
		http.Error(w, "Failed to upload photo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message":        "Photos uploaded successfully",
		"uploaded_count": len(photos),
		"photos":         buildPhotoResponses(photos),
	})
}

func storeUploadedImages(w http.ResponseWriter, r *http.Request, userID uint, formField, subDir string, maxFiles int) ([]storedUpload, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxPhotoUploadRequestSize)
	if err := r.ParseMultipartForm(maxPhotoUploadRequestSize); err != nil {
		http.Error(w, "Files are too large", http.StatusBadRequest)
		return nil, false
	}

	fileHeaders := r.MultipartForm.File[formField]
	if len(fileHeaders) == 0 {
		fileHeaders = r.MultipartForm.File["photo"]
	}
	if len(fileHeaders) == 0 {
		http.Error(w, fmt.Sprintf("Error retrieving the %s files", formField), http.StatusBadRequest)
		return nil, false
	}
	if len(fileHeaders) > maxFiles {
		http.Error(w, fmt.Sprintf("You can upload up to %d more photos", maxFiles), http.StatusBadRequest)
		return nil, false
	}

	uploads := make([]storedUpload, 0, len(fileHeaders))
	for _, fileHeader := range fileHeaders {
		if fileHeader.Size > maxUploadSize {
			removeStoredUploads(uploads)
			http.Error(w, "Each photo must be 5MB or smaller", http.StatusBadRequest)
			return nil, false
		}

		upload, ok := storeUploadedImageFromHeader(w, userID, fileHeader, subDir)
		if !ok {
			removeStoredUploads(uploads)
			return nil, false
		}
		uploads = append(uploads, upload)
	}

	return uploads, true
}

func deletePhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}

	photoID, err := strconv.ParseUint(pat.Param(r, "photoID"), 10, 64)
	if err != nil || photoID == 0 {
		http.Error(w, "Invalid photo ID", http.StatusBadRequest)
		return
	}

	var photo Photo
	if err := DB.Where("id = ? AND user_id = ?", photoID, userID).First(&photo).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			http.Error(w, "Photo not found", http.StatusNotFound)
			return
		}
		http.Error(w, "Failed to retrieve photo", http.StatusInternalServerError)
		return
	}

	if err := removeStoredUpload(photo.URL); err != nil {
		log.Printf("failed to remove photo %d for user %d: %v", photo.ID, userID, err)
		http.Error(w, "Failed to delete photo file", http.StatusInternalServerError)
		return
	}

	if err := DB.Delete(&photo).Error; err != nil {
		http.Error(w, "Failed to delete photo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]any{
		"message":  "Photo deleted successfully",
		"photo_id": photo.ID,
	})
}

func createPhotoWithinLimit(userID uint, photo *Photo) error {
	return DB.Transaction(func(tx *gorm.DB) error {
		var user User
		if err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Select("id").First(&user, userID).Error; err != nil {
			return err
		}

		if err := ensurePhotoLimitAvailable(tx, userID); err != nil {
			return err
		}

		return tx.Create(photo).Error
	})
}

func ensurePhotoLimitAvailable(db *gorm.DB, userID uint) error {
	var photoCount int64
	if err := db.Model(&Photo{}).Where("user_id = ?", userID).Count(&photoCount).Error; err != nil {
		return err
	}
	if photoCount >= maxUserPhotos {
		return errPhotoLimitReached
	}

	return nil
}

func writePhotoUploadError(w http.ResponseWriter, err error) {
	if errors.Is(err, errPhotoLimitReached) {
		http.Error(w, fmt.Sprintf("Maximum of %d photos allowed", maxUserPhotos), http.StatusBadRequest)
		return
	}
	if errors.Is(err, gorm.ErrRecordNotFound) {
		http.Error(w, "User not found", http.StatusNotFound)
		return
	}

	http.Error(w, "Failed to upload photo", http.StatusInternalServerError)
}

func storeUploadedImage(w http.ResponseWriter, r *http.Request, userID uint, formField, subDir string) (storedUpload, bool) {
	r.Body = http.MaxBytesReader(w, r.Body, maxUploadSize)
	if err := r.ParseMultipartForm(maxUploadSize); err != nil {
		http.Error(w, "File too large", http.StatusBadRequest)
		return storedUpload{}, false
	}

	file, handler, err := r.FormFile(formField)
	if err != nil {
		http.Error(w, fmt.Sprintf("Error retrieving the %s file", formField), http.StatusBadRequest)
		return storedUpload{}, false
	}
	defer file.Close()

	if err := validateImageUpload(file, handler.Filename); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return storedUpload{}, false
	}

	ext := strings.ToLower(filepath.Ext(handler.Filename))
	newFileName := fmt.Sprintf("user_%d_%d%s", userID, time.Now().UnixNano(), ext)
	uploadDir := filepath.Join(".", "uploads", subDir)
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return storedUpload{}, false
	}

	filePath := filepath.Join(uploadDir, newFileName)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return storedUpload{}, false
	}
	defer dst.Close()

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Failed to process uploaded file", http.StatusInternalServerError)
		return storedUpload{}, false
	}
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return storedUpload{}, false
	}

	return storedUpload{
		FilePath:  filePath,
		PublicURL: fmt.Sprintf("/uploads/%s/%s", subDir, newFileName),
	}, true
}

func storeUploadedImageFromHeader(w http.ResponseWriter, userID uint, fileHeader *multipart.FileHeader, subDir string) (storedUpload, bool) {
	file, err := fileHeader.Open()
	if err != nil {
		http.Error(w, "Error retrieving the photo file", http.StatusBadRequest)
		return storedUpload{}, false
	}
	defer file.Close()

	if err := validateImageUpload(file, fileHeader.Filename); err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return storedUpload{}, false
	}

	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	newFileName := fmt.Sprintf("user_%d_%d%s", userID, time.Now().UnixNano(), ext)
	uploadDir := filepath.Join(".", "uploads", subDir)
	if err := os.MkdirAll(uploadDir, 0o755); err != nil {
		http.Error(w, "Failed to create upload directory", http.StatusInternalServerError)
		return storedUpload{}, false
	}

	filePath := filepath.Join(uploadDir, newFileName)
	dst, err := os.Create(filePath)
	if err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return storedUpload{}, false
	}
	defer dst.Close()

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		http.Error(w, "Failed to process uploaded file", http.StatusInternalServerError)
		return storedUpload{}, false
	}
	if _, err := io.Copy(dst, file); err != nil {
		http.Error(w, "Failed to save file", http.StatusInternalServerError)
		return storedUpload{}, false
	}

	return storedUpload{
		FilePath:  filePath,
		PublicURL: fmt.Sprintf("/uploads/%s/%s", subDir, newFileName),
	}, true
}

func removeStoredUploads(uploads []storedUpload) {
	for _, upload := range uploads {
		_ = os.Remove(upload.FilePath)
	}
}

func removeStoredUpload(publicURL string) error {
	relativePath := strings.TrimPrefix(publicURL, "/")
	if !strings.HasPrefix(relativePath, "uploads/") {
		return fmt.Errorf("invalid upload path: %s", publicURL)
	}

	cleanPath := filepath.Clean(relativePath)
	if cleanPath == "." || strings.HasPrefix(cleanPath, "..") || !strings.HasPrefix(cleanPath, "uploads"+string(os.PathSeparator)) {
		return fmt.Errorf("invalid upload path: %s", publicURL)
	}

	err := os.Remove(cleanPath)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}

	return err
}

func validateImageUpload(file multipart.File, filename string) error {
	ext := strings.ToLower(filepath.Ext(filename))
	if _, ok := allowedImageExtensions[ext]; !ok {
		return errors.New("unsupported image format")
	}

	header := make([]byte, 512)
	n, err := file.Read(header)
	if err != nil && !errors.Is(err, io.EOF) {
		return errors.New("failed to inspect uploaded file")
	}

	contentType := http.DetectContentType(header[:n])
	if _, ok := allowedImageContentTypes[contentType]; !ok {
		return errors.New("uploaded file is not a supported image")
	}

	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return errors.New("failed to process uploaded file")
	}

	return nil
}

func parseProfileFeedLimit(raw string) (int, error) {
	if raw == "" {
		return defaultProfileFeedLimit, nil
	}

	limit, err := strconv.Atoi(raw)
	if err != nil || limit < 1 {
		return 0, errors.New("limit must be a positive integer")
	}
	if limit > maxProfileFeedLimit {
		limit = maxProfileFeedLimit
	}

	return limit, nil
}

func buildPublicProfileResponse(user User) PublicProfileResponse {
	return PublicProfileResponse{
		ID:          user.ID,
		UserName:    user.UserName,
		FirstName:   user.FirstName,
		Gender:      user.Gender,
		Preferences: user.Preferences,
		Bio:         user.Bio,
		Interests:   user.Interests,
		Avatar:      user.Avatar,
		Photos:      buildPhotoResponses(user.Photos),
	}
}

func buildPhotoResponses(photos []Photo) []PhotoResponse {
	responses := make([]PhotoResponse, 0, len(photos))
	for _, photo := range photos {
		responses = append(responses, PhotoResponse{
			ID:  photo.ID,
			URL: photo.URL,
		})
	}

	return responses
}
