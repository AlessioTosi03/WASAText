package database

type Message struct {
	ID        int    `json:"id"`
	Username  string `json:"username"`
	ConvoID   int    `json:"convo_id"`
	Text      string `json:"text"`
	Pic       string `json:"pic"`
	Forwarded bool   `json:"forwarded"`
}

func (db *appdbimpl) SendMessage(userID int, conversationID int, message string, picture string) error {
	_, err := db.c.Exec("INSERT INTO messages (user_id, conversation_id, message_text, image) VALUES (?, ?, ?, ?)", userID, conversationID, message, picture)
	return err
}

func (db *appdbimpl) ForwardMessage(userID int, conversationID int, forwardedID int) (Message, error) {
	var message Message
	err := db.c.QueryRow("SELECT message_text, image FROM messages WHERE id = ?", forwardedID).Scan(&message.Text, &message.Pic)
	if err != nil {
		return Message{}, err // Handle no user found or other errors
	}
	_, err = db.c.Exec("INSERT INTO messages (user_id, conversation_id, message_text, image, forwarded) VALUES (?, ?, ?, ?, 1)", userID, conversationID, message.Text, message.Pic)
	return message, err
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
	rowsM, errM := db.c.Query("SELECT id, user_id, message_text, image FROM messages WHERE conversation_id = ?", conversationID)
	if errM != nil {
		return nil, errM
	}
	defer rowsM.Close()
	var allMessages []Message
	for rowsM.Next() {
		var m Message
		var UserID int
		var username string
		if err := rowsM.Scan(&m.ID, &UserID, &m.Text, &m.Pic); err != nil {
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
