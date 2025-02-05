package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/julienschmidt/httprouter"
)

type Group struct {
	ID      int    `json:"id"`
	ConvoID int    `json:"convo_id"`
	Name    string `json:"name"`
	Photo   string `json:"photo"`
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

	if err := rt.db.SetGroupPhoto(group.ID, req.Photo); err != nil {
		rt.baseLogger.Errorf("Failed to update group photo for groupID %d: %v", group.ID, err)
		http.Error(w, "Failed to update group photo", http.StatusInternalServerError)
		return
	}

	group.Photo = req.Photo

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(group); err != nil {
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
		chat, err := rt.db.GetChatFromConversation(conversationID)
		if err != nil {
			rt.baseLogger.Errorf("Error retrieving chat for conversation ID %d: %v", conversationID, err)
			http.Error(w, "Internal server error", http.StatusInternalServerError)
			return
		}

		// Return chat data and messages
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := map[string]interface{}{
			"conversation": conversation,
			"chat":         chat,
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
	for _, conversation := range conversations {
		if conversation.Type == "chat" {
			chat, err := rt.db.GetChatFromConversation(conversation.ID)
			if err != nil {
				rt.baseLogger.Errorf("Error retrieving chat for conversation ID %d: %v", conversation.ID, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Return chat data and messages
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"conversation": conversation,
				"chat":         chat,
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				rt.baseLogger.Errorf("Failed to encode response: %v", err)
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}

		} else if conversation.Type == "group" {
			group, err := rt.db.GetGroupFromConversation(conversation.ID)
			if err != nil {
				rt.baseLogger.Errorf("Error retrieving group for conversation ID %d: %v", conversation.ID, err)
				http.Error(w, "Internal server error", http.StatusInternalServerError)
				return
			}

			// Return group data and messages
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := map[string]interface{}{
				"conversation": conversation,
				"group":        group,
			}
			if err := json.NewEncoder(w).Encode(response); err != nil {
				rt.baseLogger.Errorf("Failed to encode response: %v", err)
				http.Error(w, "Failed to encode response", http.StatusInternalServerError)
				return
			}
		}
	}
}
