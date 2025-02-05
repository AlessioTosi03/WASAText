package api

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/AlessioTosi03/WASAText/service/database"
	"github.com/julienschmidt/httprouter"
)

func (rt *_router) sendMessages(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	var req database.Message
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

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

	err = rt.db.SendMessage(authID, conversationID, req.Text, req.Pic)
	if err != nil {
		rt.baseLogger.Errorf("Failed to send message for userID %d: %v", authID, err)
		usererror := "Failed to send message"
		http.Error(w, usererror, http.StatusInternalServerError)
		return
	}

	message := database.Message{
		UserID:  authID,
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

	req, err = rt.db.ForwardMessage(authID, conversationID, req.ID)
	if err != nil {
		rt.baseLogger.Errorf("Failed to send message for userID %d: %v", authID, err)
		usererror := "Failed to send message"
		http.Error(w, usererror, http.StatusInternalServerError)
		return
	}

	message := database.Message{
		UserID:    authID,
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

	messageIDStr := ps.ByName("MessageID")
	messageID, _ := strconv.Atoi(messageIDStr)

	userID, err := rt.db.GetMessageUser(messageID)
	if err != nil {
		rt.baseLogger.Errorf("Failed to get message user for messageID %d: %v", messageID, err)
		http.Error(w, "Failed to get message user", http.StatusInternalServerError)
		return
	}

	if userID != authID {
		rt.baseLogger.Errorf("User %d is not the owner of message %d", authID, messageID)
		http.Error(w, "User is not the owner of message", http.StatusUnauthorized)
		return
	}

	err = rt.db.DeleteMessage(messageID)
	if err != nil {
		rt.baseLogger.Errorf("Failed to delete message for messageID %d: %v", messageID, err)
		http.Error(w, "Failed to delete message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Message deleted successfully"}`))
}

func (rt *_router) commentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageIDStr := ps.ByName("MessageID")
	messageID, _ := strconv.Atoi(messageIDStr)

	var req struct {
		Comment string `json:"comment"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

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

	err = rt.db.CommentMessage(authID, messageID, req.Comment)
	if err != nil {
		rt.baseLogger.Errorf("Failed to comment message for messageID %d: %v", messageID, err)
		http.Error(w, "Failed to comment message", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Message commented successfully"}`))
}

func (rt *_router) uncommentMessage(w http.ResponseWriter, r *http.Request, ps httprouter.Params) {
	messageIDStr := ps.ByName("MessageID")
	messageID, _ := strconv.Atoi(messageIDStr)

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

	var rowsAffected int64
	rowsAffected, err = rt.db.UncommentMessage(authID, messageID)
	if err != nil {
		rt.baseLogger.Errorf("Failed to uncomment message for messageID %d: %v", messageID, err)
		http.Error(w, "Failed to uncomment message", http.StatusInternalServerError)
		return
	}

	if rowsAffected == 0 {
		rt.baseLogger.Warnf("No comment found for messageID %d by userID %d", messageID, authID)
		http.Error(w, "Comment not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	w.Write([]byte(`{"message": "Message uncommented successfully"}`))
}
