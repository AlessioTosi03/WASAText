package database

type Conversation struct {
	ID   int    `json:"id"`
	Type string `json:"type"`
}
type Group struct {
	ID      int    `json:"id"`
	ConvoID int    `json:"convo_id"`
	Name    string `json:"name"`
	Photo   string `json:"photo"`
}
type Chat struct {
	ID      int `json:"id"`
	ConvoID int `json:"convo_id"`
}

func (db *appdbimpl) GetConversationsByUser(userID int) ([]Conversation, error) {
	var conversations []Conversation
	rows, err := db.c.Query("SELECT id, type FROM conversations WHERE id IN (SELECT conversation_id FROM participant_relation WHERE user_id = ?)", userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var conversation Conversation
		err := rows.Scan(&conversation.ID, &conversation.Type)
		if err != nil {
			return nil, err
		}
		conversations = append(conversations, conversation)
	}

	return conversations, nil
}

func (db *appdbimpl) GetGroupFromConversation(conversationID int) (Group, error) {
	var group Group
	err := db.c.QueryRow("SELECT id, group_name, group_pic, conversation_id FROM groups WHERE id = ?", conversationID).
		Scan(&group.ID, &group.Name, &group.Photo, &group.ConvoID)
	if err != nil {
		return Group{}, err // Return an empty struct and the error
	}

	return group, nil
}

func (db *appdbimpl) GetChatFromConversation(conversationID int) (Chat, error) {
	var chat Chat
	err := db.c.QueryRow("SELECT id, conversation_id FROM chats WHERE conversation_id = ?", conversationID).
		Scan(&chat.ID, &chat.ConvoID)
	if err != nil {
		return Chat{}, err // Return an empty struct and the error
	}

	return chat, nil
}

func (db *appdbimpl) GetOtherParticipant(conversationID int, userID int) (string, error) {
	var otherUserID int
	err := db.c.QueryRow("SELECT user_id FROM participant_relation WHERE conversation_id = ? AND user_id != ?", conversationID, userID).
		Scan(&otherUserID)
	if err != nil {
		return "", err // Return an empty struct and the error
	}
	var otherUsername string
	err = db.c.QueryRow("SELECT name FROM users WHERE id = ?", otherUserID).Scan(&otherUsername)
	if err != nil {
		return "", err // Return an empty struct and the error
	}

	return otherUsername, nil
}

func (db *appdbimpl) SetGroupName(groupID int, groupName string) error {
	// Attempt to update the group's name
	_, err := db.c.Exec("UPDATE groups SET group_name = ? WHERE id = ?", groupName, groupID)
	return err
}

func (db *appdbimpl) SetGroupPhoto(groupID int, photoURL string) error {
	// Attempt to update the group's photo
	_, err := db.c.Exec("UPDATE groups SET group_pic = ? WHERE id = ?", photoURL, groupID)
	return err
}

func (db *appdbimpl) GetConversation(conversationID int) (Conversation, error) {
	var conversation Conversation
	err := db.c.QueryRow("SELECT id, type FROM conversations WHERE id = ?", conversationID).
		Scan(&conversation.ID, &conversation.Type)
	if err != nil {
		return Conversation{}, err // Return an empty struct and the error
	}

	return conversation, nil
}
