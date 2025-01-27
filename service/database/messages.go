package database

type Message struct {
	ID        int    `json:"id"`
	UserID    int    `json:"user_id"`
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

func (db *appdbimpl) CommentMessage(messageID int, comment string) error {
	_, err := db.c.Exec("INSERT INTO reactions (message_id, emoji) VALUES (?, ?)", messageID, comment)
	return err
}

func (db *appdbimpl) UncommentMessage(messageID int) error {
	_, err := db.c.Exec("DELETE FROM reactions WHERE message_id = ?", messageID)
	return err
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
		if err := rowsM.Scan(&m.ID, &m.UserID, &m.Text, &m.Pic); err != nil {
			return nil, err
		}
		allMessages = append(allMessages, m)
	}
	return allMessages, nil
}
