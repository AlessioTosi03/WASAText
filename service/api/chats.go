package api

import (
	"encoding/json"
	"net/http"
	"strconv"

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

	conversationIDStr := ps.ByName("ConversationID")
	conversationID, _ := strconv.Atoi(conversationIDStr)
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

	conversationIDStr := ps.ByName("ConversationID")
	conversationID, _ := strconv.Atoi(conversationIDStr)
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
	// Extract the conversation ID from the URL parameters
	conversationIDStr := ps.ByName("ConversationID")
	conversationID, _ := strconv.Atoi(conversationIDStr)

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
