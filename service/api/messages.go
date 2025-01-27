package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/julienschmidt/httprouter"
)

type Message struct {
	ID      int    `json:"id"`
	UserID  int    `json:"user_id"`
	ConvoID int    `json:"convo_id"`
	Text    string `json:"text"`
	Pic     string `json:"pic"`
}

func (rt *_router) sendMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req Message
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

	message := Message{
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
