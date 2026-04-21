package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

const maxUploadSize = 5 << 20
const defaultProfileFeedLimit = 20
const maxProfileFeedLimit = 50

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

	if result := DB.Model(&User{}).Where("id = ?", userID).Update("avatar", upload.PublicURL); result.Error != nil {
		_ = os.Remove(upload.FilePath)
		http.Error(w, "Failed to upload avatar", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"message":    "Avatar uploaded successfully",
		"avatar_url": upload.PublicURL,
	})
}

func uploadPhoto(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(userIDKey).(uint)
	if !ok {
		http.Error(w, "Unauthorized: Invalid context", http.StatusUnauthorized)
		return
	}

	upload, ok := storeUploadedImage(w, r, userID, "photo", "photos")
	if !ok {
		return
	}

	photo := Photo{
		UserID: userID,
		URL:    upload.PublicURL,
	}
	if err := DB.Create(&photo).Error; err != nil {
		_ = os.Remove(upload.FilePath)
		http.Error(w, "Failed to upload photo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	json.NewEncoder(w).Encode(map[string]any{
		"message":   "Photo uploaded successfully",
		"photo_id":  photo.ID,
		"photo_url": photo.URL,
	})
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
