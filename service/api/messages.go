package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/AlessioTosi03/WASAText/service/database"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) sendMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req database.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	/*username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}*/
	username := "greg"
	userID, err := rt.db.GetUserIDByUsername(username)
	if err != nil {
		rt.baseLogger.Errorf("User not found for username: %s", username)
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}

	conversationIDStr := ps.ByName("ConversationID")
	conversationID, _ := strconv.Atoi(conversationIDStr)

	err = rt.db.SendMessage(userID, conversationID, req.Text, req.Pic)
	if err != nil {
		rt.baseLogger.Errorf("Failed to send message for userID %d: %v", userID, err)
		usererror := "Failed to send message"
		http.Error(w, usererror, http.StatusInternalServerError)
		return
	}

	message := database.Message{
		UserID:  userID,
		ConvoID: conversationID,
		Text:    req.Text,
		Pic:     req.Pic,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(message); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (rt *_router) forwardMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req database.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	/*username, ok := r.Context().Value("username").(string)
	if !ok || username == "" {
		http.Error(w, `{"error": "Unauthorized"}`, http.StatusUnauthorized)
		return
	}*/
	username := "greg"
	userID, err := rt.db.GetUserIDByUsername(username)
	if err != nil {
		rt.baseLogger.Errorf("User not found for username: %s", username)
		http.Error(w, `{"error": "User not found"}`, http.StatusNotFound)
		return
	}
	conversationIDStr := ps.ByName("ConversationID")
	conversationID, _ := strconv.Atoi(conversationIDStr)

	req, err = rt.db.ForwardMessage(userID, conversationID, req.ID)
	if err != nil {
		rt.baseLogger.Errorf("Failed to send message for userID %d: %v", userID, err)
		usererror := "Failed to send message"
		http.Error(w, usererror, http.StatusInternalServerError)
		return
	}

	message := database.Message{
		UserID:    userID,
		ConvoID:   conversationID,
		Text:      req.Text,
		Pic:       req.Pic,
		Forwarded: true,
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	if err := json.NewEncoder(w).Encode(message); err != nil {
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
	}
}

func (rt *_router) deleteMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageIDStr := ps.ByName("MessageID")
	messageID, _ := strconv.Atoi(messageIDStr)

	err := rt.db.DeleteMessage(messageID)
	if err != nil {
		rt.baseLogger.Errorf("Failed to delete message for messageID %d: %v", messageID, err)
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Message deleted successfully"}`))
}
