package api

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

// User represents the user object
type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	ProfilePic string `json:"profile_pic"`
}

// setMyUserName handles the PUT request to update the username
func (rt *_router) setMyUserName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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
	// Update the username in the database
	err := rt.db.SetMyUserName(userID, req.Username)
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
	var req User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	userIDStr := ps.ByName("UserID")
	userID, _ := strconv.Atoi(userIDStr)
	// Update the username in the database
	err := rt.db.SetMyPhoto(userID, req.ProfilePic)
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
	// Extract `conversationId` from the URL parameters
	var req User
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		rt.baseLogger.Errorf("Failed to decode request body: %v", err)
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	conversationIDStr := ps.ByName("ConversationID")
	conversationID, _ := strconv.Atoi(conversationIDStr)
	//!!!!need to call getIdByUsername!!!!
	var userID int
	userID, err := rt.db.GetUserIDByUsername(req.Username)
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
	// Extract the username from the context
	username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}

	// Extract `conversationID` from the URL parameters
	conversationIDStr := ps.ByName("ConversationID")
	conversationID, _ := strconv.Atoi(conversationIDStr)

	// Get the user ID using the username
	userID, err := rt.db.GetUserIDByUsername(username)
	if err != nil {
		rt.baseLogger.Errorf("User not found for username: %s", username)
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	// Perform the 'leave group' operation
	if err := rt.db.LeaveGroup(userID, conversationID); err != nil {
		rt.baseLogger.Errorf("Failed to remove user %d from conversation %d: %v", userID, conversationID, err)
		http.Error(w, `{"error": "Failed to leave group"}`, http.StatusInternalServerError)
		return
	}

	// Log success and respond with success message
	rt.baseLogger.Infof("User %d successfully left conversation %d", userID, conversationID)
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Successfully left the group"}`))
}
