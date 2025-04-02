package database

import (
	"database/sql"
	"time"
	"errors"
)

type Message struct {
	ID             int    `json:"id"`
	Username       string `json:"username"`
	ConvoID        int    `json:"convo_id"`
	Text           string `json:"text"`
	Pic            string `json:"pic"`
	Forwarded      bool   `json:"forwarded"`
	ReceivedStatus bool   `json:"received_status"`
	ReadStatus     bool   `json:"read_status"`
}
type Reaction struct {
	UserID int    `json:"user_id"`
	Emoji  string `json:"emoji"`
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
	if errors.Is(err, sql.ErrNoRows) {
		// No reaction found for the message
		return "", nil // Returning nil for no reaction found
	}
	if err != nil {
		return "", err // Handle no user found or other errors
	}
	return reaction, nil
}

func (db *appdbimpl) GetAllReactions(messageID int) ([]Reaction, error) {
	var reactions []Reaction
	rows, err := db.c.Query("SELECT emoji,user_id FROM reaction_relation WHERE message_id = ?", messageID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var reaction Reaction
		err := rows.Scan(&reaction.Emoji, &reaction.UserID)
		if err != nil {
			return nil, err
		}
		reactions = append(reactions, reaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	return reactions, nil
}

func (db *appdbimpl) ReceiveMessages(userID int) error {
	// log.Println(userID)
	var messageIDs []int
	rows, err := db.c.Query("SELECT conversation_id FROM participant_relation WHERE user_id = ?", userID)
	if err != nil {
		return err
	}
	defer rows.Close()
	for rows.Next() {
		var conversationID int
		if err := rows.Scan(&conversationID); err != nil {
			return err
		}
		mess_rows, err := db.c.Query("SELECT id FROM messages WHERE conversation_id = ?", conversationID)
		if err != nil {
			return err
		}
		defer mess_rows.Close()

		// Collezionare gli ID dei messaggi da aggiornare
		for mess_rows.Next() {
			var id int
			if err := mess_rows.Scan(&id); err != nil {
				return err
			}
			messageIDs = append(messageIDs, id)
		}
		if err := mess_rows.Err(); err != nil {
			return err
		}
		// log.Println(messageIDs)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	// Aggiornare lo stato di ricezione per ogni messaggio
	for _, messageID := range messageIDs {
		var exists int
		// log.Println(messageID)
		err := db.c.QueryRow("SELECT COUNT(*) FROM message_received_status WHERE message_id = ? AND user_id = ?", messageID, userID).Scan(&exists)
		if err != nil {
			return err
		}
		if exists == 0 {
			// Se non esiste una riga, inserisci un nuovo record
			_, err := db.c.Exec("INSERT INTO message_received_status (message_id, user_id, read_timestamp) VALUES (?, ?, ?)", messageID, userID, time.Now())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *appdbimpl) ReadMessages(userID int, conversationID int) error {
	var messageIDs []int
	mess_rows, err := db.c.Query("SELECT id FROM messages WHERE conversation_id = ?", conversationID)
	if err != nil {
		return err
	}
	defer mess_rows.Close()

	// Collezionare gli ID dei messaggi da aggiornare
	for mess_rows.Next() {
		var id int
		if err := mess_rows.Scan(&id); err != nil {
			return err
		}
		messageIDs = append(messageIDs, id)
	}
	if err := mess_rows.Err(); err != nil {
		return err
	}
	// log.Println(messageIDs)
	// Aggiornare lo stato di ricezione per ogni messaggio
	for _, messageID := range messageIDs {
		var exists int
		// log.Println(messageID)
		err := db.c.QueryRow("SELECT COUNT(*) FROM message_read_status WHERE message_id = ? AND user_id = ?", messageID, userID).Scan(&exists)
		if err != nil {
			return err
		}
		if exists == 0 {
			// Se non esiste una riga, inserisci un nuovo record
			_, err := db.c.Exec("INSERT INTO message_read_status (message_id, user_id, read_timestamp) VALUES (?, ?, ?)", messageID, userID, time.Now())
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func (db *appdbimpl) GetMessageReceivedStatus(messageID int, userID int) (int, error) {
	var UserID int
	err := db.c.QueryRow("SELECT user_id FROM message_received_status WHERE message_id = ? AND user_id = ?", messageID, userID).Scan(&UserID)
	if !errors.Is(err, sql.ErrNoRows) {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil // No read status found
		}
		return 0, err // Handle other errors
	}
	return UserID, nil
}

func (db *appdbimpl) GetMessageReadStatus(messageID int, userID int) (int, error) {
	var UserID int
	err := db.c.QueryRow("SELECT user_id FROM message_read_status WHERE message_id = ? AND user_id = ?", messageID, userID).Scan(&UserID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return 0, nil // No read status found
		}
		return 0, err // Handle other errors
	}
	return UserID, nil
}

func (db *appdbimpl) GetMessages(conversationID int, participants []string) ([]Message, error) {
	rowsM, errM := db.c.Query("SELECT id, user_id, message_text, image, forwarded, conversation_id FROM messages WHERE conversation_id = ?", conversationID)
	if errM == sql.ErrNoRows {
		// No reaction found for the message
		return []Message{}, nil // Returning nil for no reaction found
	}
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
		m.ReceivedStatus = true
		m.ReadStatus = true
		// Get received and read status
		for _, participant := range participants {
			var recUserID int
			var readUserID int
			userID, err := db.GetUserIDByUsername(participant)
			recUserID, err = db.GetMessageReceivedStatus(m.ID, userID)
			if err != nil {
				return nil, err
			}
			if recUserID == 0 {
				m.ReceivedStatus = false
			}
			readUserID, err = db.GetMessageReadStatus(m.ID, userID)
			if err != nil {
				return nil, err
			}
			if readUserID == 0 {
				m.ReadStatus = false
			}
		}
		allMessages = append(allMessages, m)
	}
	if err := rowsM.Err(); err != nil {
		return nil, err
	}
	return allMessages, nil
}
