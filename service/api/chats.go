package api

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/julienschmidt/httprouter"
)

type Group struct {
	ID      int    `json:"id"`
	ConvoID int    `json:"convo_id"`
	Name    string `json:"name"`
	Photo   string `json:"photo"`
}
type ChatUsers struct {
	ID      int    `json:"id"`
	Chatter string `json:"chatter"`
}

func (rt *_router) createGroup(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	// Parse the multipart form
	err := r.ParseMultipartForm(10 << 20) // 10 MB limit
	if err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Get name from form data
	name := r.FormValue("name")
	if name == "" {
		http.Error(w, "Group name is required", http.StatusBadRequest)
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
	if file != nil {
		// Create a unique file name
		photoPath = fmt.Sprintf("/photos/%d_%s", time.Now().Unix(), handler.Filename)

		// Save the file
		dst, err := os.Create("/home/aletos/WASAText/webui/public" + photoPath) // Save it in the server
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

	// Extract Authorization header
	authHeader := r.Header.Get("Authorization")
	if authHeader == "" {
		http.Error(w, "Missing Authorization header", http.StatusUnauthorized)
		return
	}

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

	// Store group in the database
	groupID, err := rt.db.CreateGroup(authID, name, photoPath)
	if err != nil {
		rt.baseLogger.Errorf("Failed to create group: %v", err)
		http.Error(w, "Failed to create group", http.StatusInternalServerError)
		return
	}

	// Create response
	resp := Group{
		ID:    groupID,
		Name:  name,
		Photo: photoPath,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(resp); err != nil {
		rt.baseLogger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (rt *_router) createChat(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req ChatUsers
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

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
	var otherUserID int
	otherUserID, err = rt.db.GetUserIDByUsername(req.Chatter)
	if err != nil {
		rt.baseLogger.Errorf("Failed to get user ID: %v", err)
		http.Error(w, "Failed to get user ID", http.StatusInternalServerError)
		return
	}

	chatID, err := rt.db.CreateChat(authID, otherUserID)
	if err != nil {
		rt.baseLogger.Errorf("Failed to create chat: %v", err)
		http.Error(w, "Failed to create chat", http.StatusInternalServerError)
		return
	}

	req.ID = chatID

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(w).Encode(req); err != nil {
		rt.baseLogger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (rt *_router) setGroupName(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req Group
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

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

	group, err := rt.db.GetGroupFromConversation(conversationID)
	if err != nil {
		// Check if the error indicates "no rows"
		if err.Error() == "no rows in result set" { // Error message depends on your database driver
			rt.baseLogger.Warnf("No group found for conversation ID: %d", conversationID)
			http.Error(w, "Group not found", http.StatusNotFound)
		} else {
			rt.baseLogger.Errorf("Error retrieving group for conversation ID %d: %v", conversationID, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if err := rt.db.SetGroupName(group.ID, req.Name); err != nil {
		rt.baseLogger.Errorf("Failed to update group name for groupID %d: %v", group.ID, err)
		http.Error(w, "Failed to update group name", http.StatusInternalServerError)
		return
	}

	group.Name = req.Name

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(group); err != nil {
		rt.baseLogger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) setGroupPhoto(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// Handle file upload for profile picture
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
	if file != nil {
		// Create a unique file name for the user's profile picture
		photoPath = fmt.Sprintf("/photos/%d_%s", time.Now().Unix(), handler.Filename)

		// Save the file to the server
		dst, err := os.Create("/home/aletos/WASAText/webui/public" + photoPath) // Save it in the server's public directory
		if err != nil {
			rt.baseLogger.Errorf("Error saving file: %v", err)
			http.Error(w, "Error saving file", http.StatusInternalServerError)
			return
		}
		defer dst.Close()

		// Copy the file content to the destination
		if _, err = io.Copy(dst, file); err != nil {
			http.Error(w, "Error writing file", http.StatusInternalServerError)
			return
		}
	}

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

	group, err := rt.db.GetGroupFromConversation(conversationID)
	if err != nil {
		// Check if the error indicates "no rows"
		if err.Error() == "no rows in result set" { // Error message depends on your database driver
			rt.baseLogger.Warnf("No group found for conversation ID: %d", conversationID)
			http.Error(w, "Group not found", http.StatusNotFound)
		} else {
			rt.baseLogger.Errorf("Error retrieving group for conversation ID %d: %v", conversationID, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	if err := rt.db.SetGroupPhoto(group.ID, photoPath); err != nil {
		rt.baseLogger.Errorf("Failed to update group photo for groupID %d: %v", group.ID, err)
		http.Error(w, "Failed to update group photo", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	response := map[string]interface{}{
		"photo_path": photoPath,
	}
	if err := json.NewEncoder(w).Encode(response); err != nil {
		rt.baseLogger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}
}

func (rt *_router) getConversation(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
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

	// Extract the conversation ID from the URL parameters
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

	// Fetch the conversation details from the database
	conversation, err := rt.db.GetConversation(conversationID)
	if err != nil {
		if err.Error() == "no rows in result set" {
			rt.baseLogger.Warnf("No conversation found for conversation ID: %d", conversationID)
			http.Error(w, "Conversation not found", http.StatusNotFound)
		} else {
			rt.baseLogger.Errorf("Error retrieving conversation for conversation ID %d: %v", conversationID, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
		}
		return
	}

	// Fetch the messages for the conversation
	allMessages, err := rt.db.GetMessages(conversationID)
	if err != nil {
		rt.baseLogger.Errorf("Error retrieving messages for conversation ID %d: %v", conversationID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	// Depending on the conversation type, fetch additional data (chat or group)
	if conversation.Type == "chat" {
		otherUserID, err := rt.db.GetOtherParticipant(conversationID, authID)
		if err != nil {
			rt.baseLogger.Errorf("Error retrieving other participant for conversation ID %d: %v", conversationID, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}
		// Return chat data and messages
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"conversation": conversation,
			"chatter":      otherUserID,
			"messages":     allMessages,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			rt.baseLogger.Errorf("Failed to encode response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}

	} else if conversation.Type == "group" {
		group, err := rt.db.GetGroupFromConversation(conversationID)
		if err != nil {
			rt.baseLogger.Errorf("Error retrieving group for conversation ID %d: %v", conversationID, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return group data and messages
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"conversation": conversation,
			"group":        group,
			"messages":     allMessages,
		}
		if err := json.NewEncoder(w).Encode(response); err != nil {
			rt.baseLogger.Errorf("Failed to encode response: %v", err)
			http.Error(w, "Failed to encode response", http.StatusInternalServerError)
			return
		}
	}
}

func (rt *_router) getMyConversations(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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

	// Fetch the list of conversations the user is part of
	conversations, err := rt.db.GetConversationsByUser(authID)
	if err != nil {
		rt.baseLogger.Errorf("Failed to retrieve conversations for user %d: %v", authID, err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
		return
	}

	var responseData []map[string]interface{}

	for _, conversation := range conversations {
		responseItem := map[string]interface{}{
			"conversation": conversation,
		}

		if conversation.Type == "chat" {
			chat, err := rt.db.GetChatFromConversation(conversation.ID)
			if err != nil {
				rt.baseLogger.Errorf("Error retrieving chat for conversation ID %d: %v", conversation.ID, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			otherParticipant, err := rt.db.GetOtherParticipant(conversation.ID, authID)
			if err != nil {
				rt.baseLogger.Errorf("Error retrieving other participant for conversation ID %d: %v", conversation.ID, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if otherParticipant.ProfilePic == "default" {
				otherParticipant.ProfilePic = "/photos/default.png"
			}
			responseItem["chat"] = chat
			responseItem["other_user"] = otherParticipant

		} else if conversation.Type == "group" {
			group, err := rt.db.GetGroupFromConversation(conversation.ID)
			if err != nil {
				rt.baseLogger.Errorf("Error retrieving group for conversation ID %d: %v", conversation.ID, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}
			if group.Photo == "default" {
				group.Photo = "/photos/default.png"
			}
			responseItem["group"] = group
			responseItem["other_user"] = nil
		}
		responseData = append(responseData, responseItem)
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	if err := json.NewEncoder(w).Encode(responseData); err != nil {
		rt.baseLogger.Errorf("Failed to encode response: %v", err)
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}
