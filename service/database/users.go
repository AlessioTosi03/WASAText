package database

func (db *appdbimpl) GetUserIDByUsername(username string) (int, error) {
	var userID int
	err := db.c.QueryRow("SELECT id FROM users WHERE name = ?", username).Scan(&userID)
	if err != nil {
		return 0, err // Handle no user found or other errors
	}
	return userID, nil
}

// SetName updates the username for a given user ID
func (db *appdbimpl) SetMyUserName(userID int, name string) error {
	_, err := db.c.Exec("UPDATE users SET name = ? WHERE id = ?", name, userID)
	return err
}

func (db *appdbimpl) SetMyPhoto(userID int, profile_pic string) error {
	_, err := db.c.Exec("UPDATE users SET profile_pic = ? WHERE id = ?", profile_pic, userID)
	return err
}

func (db *appdbimpl) AddToGroup(userID int, groupID int) error {
	_, err := db.c.Exec("INSERT INTO participant_relation (user_id, conversation_id) VALUES (?, ?)", userID, groupID)
	return err
}

func (db *appdbimpl) LeaveGroup(userID int, groupID int) error {
	// Attempt to remove the user from the group
	_, err := db.c.Exec("DELETE FROM participant_relation WHERE user_id = ? AND conversation_id = ?", userID, groupID)
	return err
}
