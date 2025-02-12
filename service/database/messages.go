package database

import (
	"database/sql"
)

type Message struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	ConvoID   int    `json:"convo_id"`
	Text      string `json:"text"`
	Pic       string `json:"pic"`
	Forwarded bool   `json:"forwarded"`
}

func (db *appdbimpl) SendMessage(userID int, conversationID int, message string, picture string) error {
	_, err := db.c.Exec("INSERT INTO messages (user_id, conversation_id, message_text, image, forwarded) VALUES (?, ?, ?, ?, 0)", userID, conversationID, message, picture)
	return err
}

func (db *appdbimpl) ForwardMessage(userID int, conversationID int, forwardedID int) error {
	var message Message
	err := db.c.QueryRow("SELECT message_text, image, user_id FROM messages WHERE id = ?", forwardedID).Scan(&message.Text, &message.Pic, &message.Username)
	if err != nil {
		return err // Handle no user found or other errors
	}
	_, err = db.c.Exec("INSERT INTO messages (user_id, conversation_id, message_text, image, forwarded) VALUES (?, ?, ?, ?, 1)", userID, conversationID, message.Text, message.Pic)
	return err
}

func (db *appdbimpl) DeleteMessage(messageID int) error {
	_, err := db.c.Exec("DELETE FROM messages WHERE id = ?", messageID)
	return err
}

func (db *appdbimpl) CommentMessage(userID int, messageID int, comment string) error {
	_, err := db.c.Exec("INSERT INTO reaction_relation (user_id, message_id, emoji) VALUES (?, ?, ?)", userID, messageID, comment)
	return err
}

func (db *appdbimpl) UncommentMessage(userID int, messageID int) (int64, error) {
	result, err := db.c.Exec("DELETE FROM reaction_relation WHERE user_id = ? AND message_id = ?", userID, messageID)
	if err != nil {
		return 0, err
	}
	return result.RowsAffected()
}

func (db *appdbimpl) GetMessages(conversationID int) ([]Message, error) {
	rowsM, errM := db.c.Query("SELECT id, user_id, message_text, image, forwarded, conversation_id FROM messages WHERE conversation_id = ?", conversationID)
	if errM != nil {
		return nil, errM
	}
	defer rowsM.Close()
	var allMessages []Message
	for rowsM.Next() {
		var m Message
		var UserID int
		var username string
		if err := rowsM.Scan(&m.ID, &UserID, &m.Text, &m.Pic, &m.Forwarded, &m.ConvoID); err != nil {
			return nil, err
		}
		err := db.c.QueryRow("SELECT name FROM users WHERE id = ?", UserID).Scan(&username)
		if err != nil {
			return nil, err
		}
		m.Username = username
		allMessages = append(allMessages, m)
	}
	return allMessages, nil
}

func (db *appdbimpl) GetUserByReaction(reactionID int) (int, error) {
	var userID int
	err := db.c.QueryRow("SELECT user_id FROM reaction_relation WHERE id = ?", reactionID).Scan(&userID)
	if err != nil {
		return 0, err // Handle no user found or other errors
	}
	return userID, nil
}

func (db *appdbimpl) GetMessageUser(messageID int) (int, error) {
	var userID int
	err := db.c.QueryRow("SELECT user_id FROM messages WHERE id = ?", messageID).Scan(&userID)
	if err != nil {
		return 0, err // Handle no user found or other errors
	}
	return userID, nil
}

func (db *appdbimpl) GetMyReaction(userID int, messageID int) (string, error) {
	var reaction string
	err := db.c.QueryRow("SELECT emoji FROM reaction_relation WHERE user_id = ? AND message_id = ?", userID, messageID).Scan(&reaction)
	if err == sql.ErrNoRows {
		// No reaction found for the message
		return "", nil // Returning nil for no reaction found
	}
	if err != nil {
		return "", err // Handle no user found or other errors
	}
	return reaction, nil
}
