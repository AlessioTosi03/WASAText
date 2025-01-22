package database

type Conversation struct {
	ID       int       `json:"id"`
	Type     string    `json:"type"`
	Messages []Message `json:"messages"`
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
	err := db.c.QueryRow("SELECT id, name, photo FROM conversations WHERE id = ?", conversationID).
		Scan(&conversation.ID, &conversation.Type)
	if err != nil {
		return Conversation{}, err // Return an empty struct and the error
	}

	return conversation, nil
}
