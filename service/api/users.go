package api

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

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

	// Check if the user exists
	var userID int
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

	// Return the user ID as the authentication token
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]int{"user_id": userID})
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
		rt.baseLogger.Errorf("Failed to update username for userID %d: %v", userID, err)
		usererror := fmt.Sprintf("Failed to update username: %s %d", req.Username, userID)
		http.Error(w, usererror, http.StatusInternalServerError)

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

	var req User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	userIDStr := ps.ByName("UserID")
	userID, _ := strconv.Atoi(userIDStr)

	if authID != userID {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	// Update the username in the database
	err = rt.db.SetMyPhoto(userID, req.ProfilePic)
	if err != nil {
		rt.baseLogger.Errorf("Failed to update profile pic for userID %d: %v", userID, err)
		usererror := fmt.Sprintf("Failed to update profile pic: %s %d", req.Username, userID)
		http.Error(w, usererror, http.StatusInternalServerError)
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
		json.NewEncoder(w).Encode("User not found")
		return
	}
	if err := rt.db.AddToGroup(userID, conversationID); err != nil {
		rt.baseLogger.Errorf("Failed to add user %d to conversationId %d: %v", userID, conversationID, err)
		http.Error(w, fmt.Sprintf("Failed to add user %s to group", req.Username), http.StatusInternalServerError)
		return
	}

	rt.baseLogger.Infof("Users %v successfully added to conversationId %s", userID, conversationID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Users added to group successfully"}`))
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
	w.Write([]byte(`{"message": "Successfully left the group"}`))
}
