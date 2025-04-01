package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/AlessioTosi03/WASAText/service/database"

	"github.com/julienschmidt/httprouter"
)

// User represents the user object
type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	ProfilePic string `json:"profile_pic"`
}

func (rt *_router) doLogin(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req User

	// Decode JSON body
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	// Trim whitespace and check input validity
	req.Username = strings.TrimSpace(req.Username)
	if req.Username == "" {
		http.Error(w, "Username cannot be empty", http.StatusBadRequest)
		return
	}
	var userID int
	var user database.User

	// Check if the user exists
	userID, err := rt.db.GetUserIDByUsername(req.Username)
	if err == sql.ErrNoRows {
		// User doesn't exist, create a new one
		userID, err = rt.db.CreateUsername(req.Username)
		if err != nil {
			rt.baseLogger.Errorf("Failed to create user %s: %v", req.Username, err)
			http.Error(w, "Failed to create user", http.StatusInternalServerError)
			return
		}
	} else if err != nil {
		rt.baseLogger.Errorf("Database error while retrieving user %s: %v", req.Username, err)
		http.Error(w, "Database error", http.StatusInternalServerError)
		return
	}
	user, err = rt.db.DoLogin(userID)
	if err != nil {
		rt.baseLogger.Errorf("Failed to login user %s: %v", req.Username, err)
		http.Error(w, "Failed to login user", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

// setMyUserName handles the PUT request to update the username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Check if it's in the format "Bearer <user_id>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
		return
	}

	authID, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Invalid Authorization token", http.StatusUnauthorized)
		return
	}

	var req User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	if len(req.Username) < 3 || len(req.Username) > 16 {
		http.Error(w, "Username must be between 3 and 16 characters", http.StatusBadRequest)
		return
	}
	userIDStr := ps.ByName("UserID")
	userID, _ := strconv.Atoi(userIDStr)

	if authID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Update the username in the database
	err = rt.db.SetMyUserName(userID, req.Username)
	if err != nil {
		if strings.Contains(err.Error(), "is already taken") {
			http.Error(w, err.Error(), http.StatusConflict) // 👈 Ritorna 409 al client
			return
		}
		rt.baseLogger.Errorf("Failed to update username for userID %d: %v", userID, err)
		http.Error(w, `{"error: Failed to update username"}`, http.StatusInternalServerError)

		return
	}

	user := User{
		ID:         userID,
		Username:   req.Username,
		ProfilePic: req.ProfilePic,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(user); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (rt *_router) setMyPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {

	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Check if it's in the format "Bearer <user_id>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
		return
	}

	authID, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Invalid Authorization token", http.StatusUnauthorized)
		return
	}

	// dealing with the file

	err = r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Handle file upload
	file, handler, err := r.FormFile("photo")
	if err != nil && err != http.ErrMissingFile { // It's okay if no file is provided
		http.Error(w, "Error retrieving file", http.StatusInternalServerError)
		return
	}
	defer func() {
		if file != nil {
			file.Close()
		}
	}()

	var photoPath string
	var filename string
	if file != nil {
		// Create a unique file name
		photoDir := "/tmp"
		if err := os.MkdirAll(photoDir, os.ModePerm); err != nil {
			http.Error(w, "Error creating photo directory", http.StatusInternalServerError)
		}
		filename = fmt.Sprintf("%d_%s", time.Now().Unix(), handler.Filename)
		photoPath = fmt.Sprintf("%s/%s", photoDir, filename)

		// Save the file
		dst, err := os.Create(photoPath) // Save it in the server
		if err != nil {
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		if _, err = io.Copy(dst, file); err != nil {
			http.Error(w, "Error writing file", http.StatusInternalServerError)
			return
		}
	}

	userIDStr := ps.ByName("UserID")
	userID, _ := strconv.Atoi(userIDStr)

	if authID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Update the username in the database
	err = rt.db.SetMyPhoto(userID, photoPath)
	if err != nil {
		rt.baseLogger.Errorf("Failed to update profile pic for userID %d: %v", userID, err)
		usererror := fmt.Sprintf("Failed to update profile pic for userID: %d", userID)
		http.Error(w, usererror, http.StatusInternalServerError)
		return
	}
	photoPath2 := "files/" + filename
	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"photo_path": photoPath2,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (rt *_router) addToGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Check if it's in the format "Bearer <user_id>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
		return
	}

	authID, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Invalid Authorization token", http.StatusUnauthorized)
		return
	}

	var req User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.Errorf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	conversationIDStr := ps.ByName("ConversationID")
	conversationID, _ := strconv.Atoi(conversationIDStr)

	partecipation, err := rt.db.CheckUserParticipation(authID, conversationID)
	if err != nil {
		rt.baseLogger.Errorf("Failed to check user participation: %v", err)
		http.Error(w, "Failed to check user participation", http.StatusInternalServerError)
		return
	}
	if !partecipation {
		rt.baseLogger.Errorf("User %d is not part of conversation %d", authID, conversationID)
		http.Error(w, "User is not part of conversation", http.StatusUnauthorized)
		return
	}

	var userID int
	userID, err = rt.db.GetUserIDByUsername(req.Username)
	if err != nil {
		w.WriteHeader(http.StatusNotFound)
		if err := json.NewEncoder(w).Encode("User not found"); err != nil {
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		}
		return
	}
	if err := rt.db.AddToGroup(userID, conversationID); err != nil {
		rt.baseLogger.Errorf("Failed to add user %d to conversationId %d: %v", userID, conversationID, err)
		http.Error(w, fmt.Sprintf("Failed to add user %s to group", req.Username), http.StatusInternalServerError)
		return
	}

	rt.baseLogger.Infof("Users %v successfully added to conversationId %s", userID, conversationID)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(`{"message": "Users added to group successfully"}`))
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (rt *_router) leaveGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

	// Check if it's in the format "Bearer <user_id>"
	parts := strings.Split(authHeader, " ")
	if len(parts) != 2 || parts[0] != "Bearer" {
		http.Error(w, "Invalid Authorization format", http.StatusUnauthorized)
		return
	}

	authID, err := strconv.Atoi(parts[1])
	if err != nil {
		http.Error(w, "Invalid Authorization token", http.StatusUnauthorized)
		return
	}

	// Extract `conversationID` from the URL parameters
	conversationIDStr := ps.ByName("ConversationID")
	conversationID, _ := strconv.Atoi(conversationIDStr)

	// Perform the 'leave group' operation
	if err := rt.db.LeaveGroup(authID, conversationID); err != nil {
		rt.baseLogger.Errorf("Failed to remove user %d from conversation %d: %v", authID, conversationID, err)
		http.Error(w, `{"error": "Failed to leave group"}`, http.StatusInternalServerError)
		return
	}

	// Log success and respond with success message
	rt.baseLogger.Infof("User %d successfully left conversation %d", authID, conversationID)
	w.WriteHeader(http.StatusOK)
	_, err = w.Write([]byte(`{"message": "Successfully left the group"}`))
	if err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (rt *_router) getAllUsers(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var usernames []string
	var err error
	if usernames, err = rt.db.GetAllUsers(); err != nil {
		rt.baseLogger.Errorf("Failed to get all users", err)
		http.Error(w, `{"error": "Failed to get the users"}`, http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(usernames)
}

func (rt *_router) getUserIDbyUsername(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	username := ps.ByName("username")
	userID, err := rt.db.GetUserIDByUsername(username)
	if err != nil {
		rt.baseLogger.Errorf("Failed to get userID by username %s: %v", username, err)
		http.Error(w, "Failed to get userID by username", http.StatusInternalServerError)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(userID)
}
