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
type User struct {
	ID         int    `json:"id"`
	Username   string `json:"username"`
	ProfilePic string `json:"profile_pic"`
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
	err := db.c.QueryRow("SELECT id, group_name, group_pic, conversation_id FROM groups WHERE conversation_id = ?", conversationID).
		Scan(&group.ID, &group.Name, &group.Photo, &group.ConvoID)
	if err != nil {
		return Group{}, err // Return an empty struct and the error
	}
	if group.Photo == "default" {
		group.Photo = "photos/default.png"
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

func (db *appdbimpl) GetOtherParticipant(conversationID int, userID int) (User, error) {
	var otherUserID int
	err := db.c.QueryRow("SELECT user_id FROM participant_relation WHERE conversation_id = ? AND user_id != ?", conversationID, userID).
		Scan(&otherUserID)
	if err != nil {
		return User{}, err // Return an empty struct and the error
	}

	otherUser := User{otherUserID, "", ""}
	err = db.c.QueryRow("SELECT name, profile_pic FROM users WHERE id = ?", otherUserID).Scan(&otherUser.Username, &otherUser.ProfilePic)
	if err != nil {
		return User{}, err // Return an empty struct and the error
	}
	if otherUser.ProfilePic == "default" {
		otherUser.ProfilePic = "photos/default.png"
	}
	return otherUser, nil
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

func (db *appdbimpl) CreateGroup(userID int, groupName string, photoURL string) (int, error) {
	// Attempt to insert the new group
	conv, err := db.c.Exec("INSERT INTO conversations (type) VALUES ('group')")
	if err != nil {
		return 0, err
	}
	convoID, err := conv.LastInsertId()
	if err != nil {
		return 0, err
	}

	res, err := db.c.Exec("INSERT INTO groups (group_name, group_pic, conversation_id) VALUES (?, ?, ?)", groupName, photoURL, convoID)
	if err != nil {
		return 0, err
	}
	groupID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Attempt to insert the user as a participant
	_, err = db.c.Exec("INSERT INTO participant_relation (conversation_id, user_id) VALUES (?, ?)", convoID, userID)
	if err != nil {
		return 0, err
	}

	return int(groupID), nil
}

func (db *appdbimpl) CreateChat(userID int, otherUserID int) (int, error) {
	conv, err := db.c.Exec("INSERT INTO conversations (type) VALUES ('chat')")
	if err != nil {
		return 0, err
	}
	convoID, err := conv.LastInsertId()
	if err != nil {
		return 0, err
	}
	// Attempt to insert the new chat
	res, err := db.c.Exec("INSERT INTO chats (conversation_id) VALUES (?)", convoID)
	if err != nil {
		return 0, err
	}
	chatID, err := res.LastInsertId()
	if err != nil {
		return 0, err
	}

	// Attempt to insert the user as a participant
	_, err = db.c.Exec("INSERT INTO participant_relation (conversation_id, user_id) VALUES (?, ?)", convoID, userID)
	if err != nil {
		return 0, err
	}

	// Attempt to insert the other user as a participant
	_, err = db.c.Exec("INSERT INTO participant_relation (conversation_id, user_id) VALUES (?, ?)", convoID, otherUserID)
	if err != nil {
		return 0, err
	}

	return int(chatID), nil
}

func (db *appdbimpl) GetGroupParticipants(groupID int) ([]string, error) {
	var users []string
	rows, err := db.c.Query("SELECT user_id FROM participant_relation WHERE conversation_id = ?", groupID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var user string
		var userID int
		err := rows.Scan(&userID)
		if err != nil {
			return nil, err
		}
		err = db.c.QueryRow("SELECT name FROM users WHERE id = ?", userID).Scan(&user)
		if err != nil {
			return nil, err
		}
		users = append(users, user)
	}

	return users, nil
}